package rider

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"regexp"
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

	// 新增字段用于聊天功能
	MerchantID uint `json:"merchantId"`
	UserID     uint `json:"userId"`     // 订单用户ID
	UserBaseID uint `json:"userBaseId"` // 用户的base_user_id，用于聊天

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

	// 新增字段
	MerchantID uint `gorm:"column:merchant_id"`
	UserID     uint `gorm:"column:user_id"`
	UserBaseID uint `gorm:"column:user_base_id"`

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

// 计算两点之间的距离（单位：米）
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000 // 地球半径（米）

	// 将经纬度转换为弧度
	φ1 := lat1 * math.Pi / 180
	φ2 := lat2 * math.Pi / 180
	Δφ := (lat2 - lat1) * math.Pi / 180
	Δλ := (lon2 - lon1) * math.Pi / 180

	// Haversine公式
	a := math.Sin(Δφ/2)*math.Sin(Δφ/2) +
		math.Cos(φ1)*math.Cos(φ2)*
			math.Sin(Δλ/2)*math.Sin(Δλ/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return R * c
}

// 智能地址补全函数（与前端保持一致）
func enhanceAddress(address string, addressType string) string {
	if address == "" {
		return ""
	}

	originalAddress := address

	// 如果地址太简单，尝试智能补全（与前端逻辑保持一致）
	if len(address) < 5 {
		fmt.Printf("⚠️ [后端地址补全] 地址过于简单: %q，尝试智能补全\n", originalAddress)

		// 中山大学珠海校区常见地点映射
		campusLocations := map[string]string{
			"容园": "广东省珠海市香洲区中山大学珠海校区榕园",
			"榕园": "广东省珠海市香洲区中山大学珠海校区榕园",
			"荔园": "广东省珠海市香洲区中山大学珠海校区荔园",
			"食堂": "广东省珠海市香洲区中山大学珠海校区食堂",
			"宿舍": "广东省珠海市香洲区中山大学珠海校区学生宿舍",
			"教学楼": "广东省珠海市香洲区中山大学珠海校区教学楼",
			"图书馆": "广东省珠海市香洲区中山大学珠海校区图书馆",
			"超市": "广东省珠海市香洲区中山大学珠海校区超市",
		}

		// 尝试模糊匹配关键词
		for key, location := range campusLocations {
			if strings.Contains(address, key) || strings.Contains(key, address) {
				fmt.Printf("✅ [后端地址补全] 智能匹配: %q -> %q\n", originalAddress, location)
				return location
			}
		}

		// 处理数字地址（可能是楼号、宿舍号等）
		if matched, _ := regexp.MatchString(`^\d+$`, address); matched {
			enhancedAddress := fmt.Sprintf("广东省珠海市香洲区中山大学珠海校区%s栋", address)
			fmt.Printf("✅ [后端地址补全] 数字地址补全: %q -> %q\n", originalAddress, enhancedAddress)
			return enhancedAddress
		}

		// 处理"容9"这类格式（数字+文字或文字+数字）
		matched, _ := regexp.MatchString(`\d`, address)
		if matched {
			enhancedAddress := fmt.Sprintf("广东省珠海市香洲区中山大学珠海校区%s", address)
			fmt.Printf("✅ [后端地址补全] 楼栋地址补全: %q -> %q\n", originalAddress, enhancedAddress)
			return enhancedAddress
		}

		// 默认补全到中山大学珠海校区
		defaultEnhanced := "广东省珠海市香洲区中山大学珠海校区"
		fmt.Printf("⚠️ [后端地址补全] 默认补全: %q -> %q\n", originalAddress, defaultEnhanced)
		return defaultEnhanced
	}

	return originalAddress
}

// 解析地址获取经纬度（使用高德地图API）
func parseAddressToCoords(address string) (lat, lon float64, err error) {
	fmt.Printf("🌍 [parseAddressToCoords] 输入地址: %q (长度:%d)\n", address, len(address))

	if address == "" {
		fmt.Printf("❌ [parseAddressToCoords] 地址为空\n")
		return 0, 0, errors.New("地址为空")
	}

	// 智能补全地址（与前端保持一致）
	enhancedAddress := enhanceAddress(address, "delivery")
	fmt.Printf("🔧 [parseAddressToCoords] 智能补全后地址: %q\n", enhancedAddress)

	fmt.Printf("🔍 [parseAddressToCoords] 调用 utils.GeoCode 解析地址\n")

	// 调用高德地图地理编码API
	lng, lat, err := utils.GeoCode(enhancedAddress)
	if err != nil {
		fmt.Printf("❌ [parseAddressToCoords] utils.GeoCode 失败: %v\n", err)
		return 0, 0, fmt.Errorf("无法解析收货地址坐标：%v，地址：%s", err, enhancedAddress)
	}

	fmt.Printf("✅ [parseAddressToCoords] 解析成功: %q -> lng=%.8f, lat=%.8f\n", enhancedAddress, lng, lat)
	fmt.Printf("📍 [parseAddressToCoords] 返回: lat=%.8f, lon=%.8f\n", lat, lng)

	return lat, lng, nil // 注意返回顺序：先纬度后经度
}

// ✅ 4) 送达：4 -> 5（需要距离校验）
// POST /api/rider/orders/:id/deliver
func DeliverOrder(c *gin.Context) {
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

	// 1. 获取骑手当前位置（强制获取最新记录）
	var riderProfile models.RiderProfile
	if err := global.Db.Where("user_id = ?", baseUserID).Order("updated_at DESC").First(&riderProfile).Error; err != nil {
		fmt.Printf("❌ [距离校验] 骑手信息查询失败: %v, baseUserID: %d\n", err, baseUserID)
		fail(c, "未获取到骑手当前位置，请先上报定位")
		return
	}

	// 检查位置数据时效性（最近10分钟内）
	if time.Since(riderProfile.UpdatedAt) > 10*time.Minute {
		fmt.Printf("❌ [距离校验] 骑手位置数据过期: 最后更新=%v, 当前=%v, 相差=%v\n",
			riderProfile.UpdatedAt, time.Now(), time.Since(riderProfile.UpdatedAt))
		fail(c, "骑手位置数据过期，请重新上报定位")
		return
	}

	// 检查骑手是否有位置信息
	if riderProfile.Latitude == 0 || riderProfile.Longitude == 0 {
		fmt.Printf("❌ [距离校验] 骑手位置无效: lat=%.8f, lng=%.8f\n", riderProfile.Latitude, riderProfile.Longitude)
		fail(c, "未获取到骑手当前位置，请先上报定位")
		return
	}

	// 检查位置是否在合理范围内（珠海地区）
	if riderProfile.Latitude < 21.5 || riderProfile.Latitude > 23.5 ||
	   riderProfile.Longitude < 112.5 || riderProfile.Longitude > 114.5 {
		fmt.Printf("❌ [距离校验] 骑手位置超出合理范围: lat=%.8f, lng=%.8f\n", riderProfile.Latitude, riderProfile.Longitude)
		fail(c, "骑手位置异常，请重新获取定位")
		return
	}

	fmt.Printf("✅ [距离校验] 骑手位置验证通过: lat=%.8f, lng=%.8f, 更新时间=%v\n",
		riderProfile.Latitude, riderProfile.Longitude, riderProfile.UpdatedAt)

	// 2. 获取订单的收货地址坐标
	type OrderInfo struct {
		DeliveryAddress string
		Province        sql.NullString
		City            sql.NullString
		District        sql.NullString
		Street          sql.NullString
		Detail          sql.NullString
	}

	var orderInfo OrderInfo
	err = global.Db.Raw(`
		SELECT
			o.delivery_address,
			a.province, a.city, a.district, a.street, a.detail
		FROM orders o
		LEFT JOIN consignees c ON c.id = o.consigneeid
		LEFT JOIN addresses a ON a.id = c.addressid
		WHERE o.id = ? AND o.rider_id = ? AND o.status = ?
	`, orderID, riderID, OrderStatusDelivering).Scan(&orderInfo).Error

	if err != nil {
		fail(c, "查询订单失败")
		return
	}

	// 如果delivery_address为空，尝试拼接address字段
	deliveryAddress := orderInfo.DeliveryAddress
	if deliveryAddress == "" {
		parts := []string{
			orderInfo.Province.String,
			orderInfo.City.String,
			orderInfo.District.String,
			orderInfo.Street.String,
			orderInfo.Detail.String,
		}
		var sb strings.Builder
		for _, p := range parts {
			if p != "" {
				sb.WriteString(p)
			}
		}
		deliveryAddress = sb.String()
	}

	if deliveryAddress == "" {
		fail(c, "无法获取订单收货地址")
		return
	}

	// 3. 解析收货地址坐标
	fmt.Printf("🗺️ [后端地址解析] 准备解析地址: %q\n", deliveryAddress)
	fmt.Printf("🏗️ [后端地址解析] 地址组件: 省份=%q, 城市=%q, 区县=%q, 街道=%q, 详情=%q\n",
		orderInfo.Province.String, orderInfo.City.String, orderInfo.District.String, orderInfo.Street.String, orderInfo.Detail.String)

	destLat, destLon, err := parseAddressToCoords(deliveryAddress)
	if err != nil {
		fmt.Printf("❌ [后端地址解析] 失败: %v\n", err)
		fail(c, err.Error())
		return
	}

	fmt.Printf("✅ [后端地址解析] 成功: %q -> (%.8f, %.8f)\n", deliveryAddress, destLat, destLon)

	// 4. 计算距离
	distance := calculateDistance(
		riderProfile.Latitude, riderProfile.Longitude,
		destLat, destLon,
	)

	// 距离阈值：100米（降低阈值，提高严格程度）
	const maxDistance = 100.0

	fmt.Printf("🚨 [距离校验] 距离检查:\n")
	fmt.Printf("   🏍️ 骑手位置: lat=%.8f, lng=%.8f\n", riderProfile.Latitude, riderProfile.Longitude)
	fmt.Printf("   📍 目标位置: lat=%.8f, lng=%.8f\n", destLat, destLon)
	fmt.Printf("   📏 计算距离: %.2f米\n", distance)
	fmt.Printf("   ⚠️ 距离阈值: %.2f米\n", maxDistance)

	if distance > maxDistance {
		fmt.Printf("❌ [距离校验失败] 距离超出限制: %.2f > %.2f\n", distance, maxDistance)
		fail(c, fmt.Sprintf("不在收货点附近（距离约 %d米），无法确认送达", int(distance)))
		return
	}

	fmt.Printf("✅ [距离校验通过] 距离符合要求: %.2f <= %.2f\n", distance, maxDistance)

	// 5. 通过距离校验，执行送达流程
	changeStatus(c, OrderStatusDelivering, OrderStatusDone)
}

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
	case OrderStatusDone:
		updates["deliver_at"] = &now
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
  o.merchant_id, o.userid,
  u.base_id AS user_base_id,
  m.shop_name, m.shop_location,
  c.name AS customer_name,
  a.province, a.city, a.district, a.street, a.detail
FROM orders o
LEFT JOIN merchants  m ON m.id = o.merchant_id
LEFT JOIN users      u ON u.id = o.userid
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
		pickupAddr := r.ShopLocation.String
		deliveryAddr := buildAddr(r)

		// 后端调试日志
		fmt.Printf("📍 [订单地址调试] ID:%d, 状态:%d\n", r.ID, r.Status)
		fmt.Printf("  🏪 商家: %s\n", r.ShopName.String)
		fmt.Printf("  📮 pickupAddress: %q (长度:%d)\n", pickupAddr, len(pickupAddr))
		fmt.Printf("  🏠 客户: %s\n", r.CustomerName.String)
		fmt.Printf("  📍 deliveryAddress: %q (长度:%d)\n", deliveryAddr, len(deliveryAddr))
		fmt.Printf("  🏗️ 地址组件: 省=%q,市=%q,区=%q,街=%q,详=%q\n",
			r.Province.String, r.City.String, r.District.String, r.Street.String, r.Detail.String)

		list = append(list, OrderItemResp{
			ID:              r.ID,
			Restaurant:      r.ShopName.String,
			PickupAddress:   pickupAddr,
			Customer:        r.CustomerName.String,
			DeliveryAddress: deliveryAddr,
			Distance:        1.2,
			EstimatedFee:    r.DeliveryFee, // 想展示总价就改成 r.TotalPrice
			EstimatedTime:   20,
			CreatedAt:       r.CreatedAt,
			Status:          r.Status,

			// 新增字段
			MerchantID: r.MerchantID,
			UserID:     r.UserID,
			UserBaseID: r.UserBaseID,

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
