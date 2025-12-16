package rider

import (
	"backend/global"
	"backend/models"
	"database/sql"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type OrderItemResp struct {
	ID              uint      `json:"id"`
	Restaurant      string    `json:"restaurant"`
	PickupAddress   string    `json:"pickupAddress"`
	Customer        string    `json:"customer"`
	DeliveryAddress string    `json:"deliveryAddress"`
	Distance        float64   `json:"distance"`
	EstimatedFee    float64   `json:"estimatedFee"`
	EstimatedTime   int       `json:"estimatedTime"`
	CreatedAt       time.Time `json:"createdAt"`
	Status          int       `json:"status"`

	AcceptedAt *time.Time `json:"acceptedAt"`
	PickupAt   *time.Time `json:"pickupAt"`
	DeliverAt  *time.Time `json:"deliverAt"`
	FinishAt   *time.Time `json:"finishAt"`
}

// orders.rider_id 存的是 Rider 表的主键 ID（不是 base_user_id）
func getRiderIDFromBaseUser(baseUserID uint) (uint, error) {
	var r models.Rider
	if err := global.Db.Where("base_id = ?", baseUserID).First(&r).Error; err != nil {
		return 0, err
	}
	return r.ID, nil
}

type orderJoinRow struct {
	ID          uint      `gorm:"column:id"`
	Status      int       `gorm:"column:status"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	TotalPrice  float64   `gorm:"column:total_price"`
	DeliveryFee float64   `gorm:"column:delivery_fee"`

	AcceptedAt *time.Time `gorm:"column:accepted_at"`
	PickupAt   *time.Time `gorm:"column:pickup_at"`
	DeliverAt  *time.Time `gorm:"column:deliver_at"`
	FinishAt   *time.Time `gorm:"column:finish_at"`

	ShopName     sql.NullString `gorm:"column:shop_name"`
	ShopLocation sql.NullString `gorm:"column:shop_location"`
	CustomerName sql.NullString `gorm:"column:customer_name"`

	Province sql.NullString `gorm:"column:province"`
	City     sql.NullString `gorm:"column:city"`
	District sql.NullString `gorm:"column:district"`
	Street   sql.NullString `gorm:"column:street"`
	Detail   sql.NullString `gorm:"column:detail"`
}

func buildAddr(r orderJoinRow) string {
	parts := []string{r.Province.String, r.City.String, r.District.String, r.Street.String, r.Detail.String}
	var sb strings.Builder
	for _, p := range parts {
		if p != "" {
			sb.WriteString(p)
		}
	}
	return sb.String()
}

// 结算：completed_orders+1 + income_record + wallet 入账（幂等）
func settleRiderForOrder(tx *gorm.DB, baseUserID uint, riderID uint, orderID uint, amount float64) error {
	// 幂等：已经结算过就直接返回
	var cnt int64
	if err := tx.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND order_id = ? AND type = ?", riderID, orderID, "order").
		Count(&cnt).Error; err != nil {
		return err
	}
	if cnt > 0 {
		return nil
	}

	// 1) completed_orders + 1（RiderProfile 的 user_id 是 baseUserID）
	if err := tx.Model(&models.RiderProfile{}).
		Where("user_id = ?", baseUserID).
		UpdateColumn("completed_orders", gorm.Expr("completed_orders + 1")).Error; err != nil {
		return err
	}

	// 2) 插入收入流水
	rec := models.RiderIncomeRecord{
		RiderID: riderID,
		OrderID: orderID,
		Amount:  amount,
		Type:    "order",
		Remark:  "订单配送收入",
	}
	if err := tx.Create(&rec).Error; err != nil {
		return err
	}

	// 3) 钱包入账（没有就创建）
	var w models.RiderWallet
	err := tx.Where("rider_id = ?", riderID).First(&w).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w = models.RiderWallet{
				RiderID:      riderID,
				Balance:      amount,
				FrozenAmount: 0,
				TotalIncome:  amount,
			}
			return tx.Create(&w).Error
		}
		return err
	}

	return tx.Model(&models.RiderWallet{}).
		Where("rider_id = ?", riderID).
		Updates(map[string]any{
			"balance":      gorm.Expr("balance + ?", amount),
			"total_income": gorm.Expr("total_income + ?", amount),
		}).Error
}

// ✅ 1) 待接单池（骑手端 new）：status=3 且 rider_id=0
// GET /api/rider/orders/new
func GetNewOrders(c *gin.Context) {
	list, err := queryOrdersJoined(nil, []int{OrderStatusToDeliver}, 50, true)
	if err != nil {
		ok(c, make([]OrderItemResp, 0))
		return
	}
	ok(c, list)
}

// ✅ 2) 接单：不改 status（仍为3），只抢单绑定 rider_id + accepted_at
// POST /api/rider/orders/:id/accept
func AcceptOrder(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")
	orderID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, "订单ID错误")
		return
	}
	orderID := uint(orderID64)

	riderID, err := getRiderIDFromBaseUser(baseUserID)
	if err != nil {
		fail(c, "未找到骑手身份（Rider 表）")
		return
	}

	now := time.Now()
	updates := map[string]any{
		"rider_id":    riderID,
		"accepted_at": &now,
		// status 不动：仍为 3
	}

	res := global.Db.Table("orders").
		Where("id = ? AND status = ? AND rider_id = 0", orderID, OrderStatusToDeliver).
		Updates(updates)

	if res.Error != nil {
		fail(c, "更新失败")
		return
	}
	if res.RowsAffected == 0 {
		fail(c, "订单已被他人接单或状态不允许")
		return
	}

	ok(c, gin.H{"success": true})
}

// ✅ 3) 取货：3 -> 4
// POST /api/rider/orders/:id/pickup
func PickupOrder(c *gin.Context) { changeStatus(c, OrderStatusToDeliver, OrderStatusDelivering) }

// ✅ 4) 送达：4 -> 5（结算）
// POST /api/rider/orders/:id/deliver
func DeliverOrder(c *gin.Context) { changeStatus(c, OrderStatusDelivering, OrderStatusDone) }

func changeStatus(c *gin.Context, from, to int) {
	baseUserID := c.GetUint("baseUserID")
	orderID64, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		fail(c, "订单ID错误")
		return
	}
	orderID := uint(orderID64)

	riderID, err := getRiderIDFromBaseUser(baseUserID)
	if err != nil {
		fail(c, "未找到骑手身份（Rider 表）")
		return
	}

	now := time.Now()
	updates := map[string]any{"status": to}

	switch to {
	case OrderStatusDelivering:
		updates["pickup_at"] = &now
		updates["deliver_at"] = &now
	case OrderStatusDone:
		updates["finish_at"] = &now
		updates["rider_id"] = riderID // 兜底：确保历史归属
	}

	err = global.Db.Transaction(func(tx *gorm.DB) error {
		// ✅ 取货/送达 都必须属于该骑手
		res := tx.Table("orders").
			Where("id = ? AND status = ? AND rider_id = ?", orderID, from, riderID).
			Updates(updates)

		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errors.New("订单状态不允许或不属于你")
		}

		// ✅ 只有完成（4->5）才结算
		if to == OrderStatusDone {
			var fee float64
			if err := tx.Table("orders").
				Select("delivery_fee").
				Where("id = ?", orderID).
				Scan(&fee).Error; err != nil {
				return err
			}
			if err := settleRiderForOrder(tx, baseUserID, riderID, orderID, fee); err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		fail(c, err.Error())
		return
	}

	ok(c, gin.H{"success": true})
}

// ✅ 5) 进行中：status in (3,4) 且 rider_id=自己
// GET /api/rider/orders/ongoing
func GetOngoingOrders(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")
	riderID, err := getRiderIDFromBaseUser(baseUserID)
	if err != nil {
		fail(c, "未找到骑手身份（Rider 表）")
		return
	}
	list, err := queryOrdersJoined(&riderID, []int{OrderStatusToDeliver, OrderStatusDelivering}, 100, false)
	if err != nil {
		fail(c, "查询失败")
		return
	}
	ok(c, list)
}

// ✅ 6) 历史：status=5 且 rider_id=自己
// GET /api/rider/orders/history
func GetHistoryOrders(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")
	riderID, err := getRiderIDFromBaseUser(baseUserID)
	if err != nil {
		fail(c, "未找到骑手身份（Rider 表）")
		return
	}
	list, err := queryOrdersJoined(&riderID, []int{OrderStatusDone}, 100, false)
	if err != nil {
		fail(c, "查询失败")
		return
	}
	ok(c, list)
}

// riderID == nil: 不按骑手过滤（new orders）
// riderID != nil: 只看该骑手订单（ongoing/history）
// onlyUnassigned: 只看 rider_id=0（用于 new）
func queryOrdersJoined(riderID *uint, statuses []int, limit int, onlyUnassigned bool) ([]OrderItemResp, error) {
	if len(statuses) == 0 {
		return make([]OrderItemResp, 0), nil
	}

	var rows []orderJoinRow

	baseSQL := `
SELECT
  o.id, o.status, o.created_at, o.total_price, o.delivery_fee,
  o.accepted_at, o.pickup_at, o.deliver_at, o.finish_at,
  m.shop_name, m.shop_location,
  c.name AS customer_name,
  a.province, a.city, a.district, a.street, a.detail
FROM orders o
LEFT JOIN merchants  m ON m.id = o.merchant_id
LEFT JOIN consignees c ON c.id = o.consigneeid
LEFT JOIN addresses  a ON a.id = c.addressid
WHERE o.status IN ?
`

	args := []any{statuses}

	if onlyUnassigned {
		baseSQL += " AND o.rider_id = 0 "
	}

	if riderID != nil {
		baseSQL += " AND o.rider_id = ? "
		args = append(args, *riderID)
	}

	baseSQL += " ORDER BY o.created_at DESC LIMIT ? "
	args = append(args, limit)

	if err := global.Db.Raw(baseSQL, args...).Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]OrderItemResp, 0, len(rows))
	for _, r := range rows {
		list = append(list, OrderItemResp{
			ID:              r.ID,
			Restaurant:      r.ShopName.String,
			PickupAddress:   r.ShopLocation.String,
			Customer:        r.CustomerName.String,
			DeliveryAddress: buildAddr(r),
			Distance:        1.2,
			EstimatedFee:    r.DeliveryFee, // 想展示总价就改成 r.TotalPrice
			EstimatedTime:   20,
			CreatedAt:       r.CreatedAt,
			Status:          r.Status,

			AcceptedAt: r.AcceptedAt,
			PickupAt:   r.PickupAt,
			DeliverAt:  r.DeliverAt,
			FinishAt:   r.FinishAt,
		})
	}

	if len(list) == 0 {
		return make([]OrderItemResp, 0), nil
	}
	return list, nil
}
