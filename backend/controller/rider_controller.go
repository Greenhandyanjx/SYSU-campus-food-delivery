package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GET /rider/info
func GetRiderInfo(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")

	var p models.RiderProfile
	if err := global.Db.Where("user_id = ?", baseUserID).First(&p).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "未找到骑手信息"})
		return
	}

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"id":              p.UserID, // 或者 p.RiderID 看你 models 定义
			"name":            p.Name,
			"avatar":          p.Avatar,
			"phone":           p.Phone,
			"rating":          p.Rating,
			"completedOrders": p.CompletedOrders,
			"isOnline":        p.IsOnline,
		},
	})
}

// POST /rider/status
func UpdateRiderStatus(c *gin.Context) {
	var req struct {
		IsOnline bool `json:"isOnline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	baseUserID := c.GetUint("baseUserID")

	err := global.Db.Model(&models.RiderProfile{}).Where("user_id = ?", baseUserID).
		Update("is_online", req.IsOnline).Error

	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": gin.H{"success": true}})
}

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
}

func GetNewOrders(c *gin.Context) {
	type NewOrderRow struct {
		ID            uint      `json:"id"`
		ShopName      string    `json:"shop_name"`
		ShopLocation  string    `json:"shop_location"`
		ConsigneeName string    `json:"consignee_name"`
		Province      string    `json:"province"`
		City          string    `json:"city"`
		District      string    `json:"district"`
		Street        string    `json:"street"`
		Detail        string    `json:"detail"`
		TotalPrice    float64   `json:"total_price"`
		CreatedAt     time.Time `json:"created_at"`
	}

	var rows []NewOrderRow

	// 一条 SQL 把订单 + 商家 + 收货人 + 地址全联表查出来
	if err := global.Db.
		Table("orders AS o").
		Select(`o.id,
		        o.total_price,
		        o.created_at,
		        m.shop_name,
		        m.shop_location,
		        c.name         AS consignee_name,
		        a.province,
		        a.city,
		        a.district,
		        a.street,
		        a.detail`).
		Joins("JOIN merchants  AS m ON m.id = o.merchant_id").
		Joins("JOIN consignees AS c ON c.id = o.consigneeid").
		Joins("JOIN addresses  AS a ON a.id = c.addressid").
		Where("o.status = ? AND o.rider_id IS NULL", 3). // 3: 待接单
		Order("o.created_at DESC").
		Limit(50).
		Scan(&rows).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	list := make([]OrderItemResp, 0, len(rows))

	for _, r := range rows {
		fullAddr := r.Province + r.City + r.District + r.Street + r.Detail

		list = append(list, OrderItemResp{
			ID:              r.ID,
			Restaurant:      r.ShopName,
			PickupAddress:   r.ShopLocation,
			Customer:        r.ConsigneeName,
			DeliveryAddress: fullAddr,
			Distance:        1.2,          // 你们前端现在写死，保持不变
			EstimatedFee:    r.TotalPrice, // 先用订单总价当预估费用
			EstimatedTime:   20,           // 先写死 20 分钟
			CreatedAt:       r.CreatedAt,
		})
	}

	c.JSON(200, gin.H{"code": 1, "data": list})
}

// POST /rider/orders/:orderId/accept_safe
func AcceptOrderSafe(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID")

	tx := global.Db.Begin()

	var order models.Order
	if err := tx.Set("gorm:query_option", "FOR UPDATE").
		Where("id = ?", orderId).First(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在"})
		return
	}

	// 只有 status = 3 且尚未分配骑手的订单可以被抢
	if order.Status != 3 || order.RiderID != 0 {
		tx.Rollback()
		c.JSON(200, gin.H{"code": 0, "msg": "订单已被抢走"})
		return
	}

	pickupCode := utils.GeneratePickupCode()
	now := time.Now()

	// 接单后直接进入配送中（4），但此时 pickup_at 仍为空
	err := tx.Model(&order).Updates(map[string]interface{}{
		"status":      4, // 配送中
		"rider_id":    riderID,
		"pickup_code": pickupCode,
		"accepted_at": &now,
	}).Error

	if err != nil {
		tx.Rollback()
		c.JSON(200, gin.H{"code": 0, "msg": "接单失败"})
		return
	}

	tx.Commit()

	c.JSON(200, gin.H{"code": 1, "data": gin.H{
		"success":    true,
		"pickupCode": pickupCode,
	}})
}

// 骑手取货接口
func PickupOrder(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID")

	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	// 必须是骑手已接单的配送中订单（4）
	if order.Status != 4 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单状态不正确"})
		return
	}

	// 已经取过货就不重复操作
	if order.PickupAt != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单已取货"})
		return
	}

	now := time.Now()
	err := global.Db.Model(&order).Updates(map[string]interface{}{
		"pickup_at": &now,
	}).Error

	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "取货失败"})
		return
	}

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"success": true,
		},
	})
}

// GET /rider/orders/delivering
func GetDeliveringOrders(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	type DeliveringOrderResp struct {
		ID              uint   `json:"id"`
		Customer        string `json:"customer"`
		CustomerPhone   string `json:"customerPhone"`
		CustomerAvatar  string `json:"customerAvatar"`
		DeliveryAddress string `json:"deliveryAddress"`
		RemainingTime   int    `json:"remainingTime"`
	}

	var orders []models.Order
	if err := global.Db.
		Where("rider_id = ? AND status = ? AND pickup_at IS NOT NULL AND finish_at IS NULL", riderID, 4).
		Order("updated_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	list := make([]DeliveringOrderResp, 0, len(orders))

	for _, o := range orders {
		var consignee models.Consignee
		global.Db.Where("id = ?", o.Consigneeid).First(&consignee)

		var addr models.Address
		global.Db.Where("id = ?", consignee.Addressid).First(&addr)

		fullAddr := addr.Province + addr.City + addr.District + addr.Street + addr.Detail

		// 计算剩余时间（分钟），小于 0 就置 0
		remaining := 0
		if !o.ExpectedTime.IsZero() {
			diff := int(time.Until(o.ExpectedTime).Minutes())
			if diff > 0 {
				remaining = diff
			}
		}

		list = append(list, DeliveringOrderResp{
			ID:              o.ID,
			Customer:        consignee.Name,
			CustomerPhone:   consignee.Phone,
			CustomerAvatar:  "",
			DeliveryAddress: fullAddr,
			RemainingTime:   remaining,
		})
	}

	c.JSON(200, gin.H{"code": 1, "data": list})
}

func CompleteOrder(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	orderId := c.Param("orderId")

	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	// 只能在配送中（4）时完成
	if order.Status != 4 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单状态不正确"})
		return
	}

	now := time.Now()

	// ==== 1. 更新订单状态为已完成（5） ====
	if err := global.Db.Model(&models.Order{}).
		Where("id = ?", order.ID).
		Updates(map[string]interface{}{
			"status":       5, // 已完成
			"dropof_point": now,
			"finish_at":    now,
		}).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "完成配送失败"})
		return
	}

	// ==== 2. 写入收入记录 ====
	global.Db.Create(&models.RiderIncomeRecord{
		RiderID:   riderID,
		OrderID:   order.ID,
		Amount:    order.TotalPrice,
		Type:      "delivery",
		Remark:    "配送完成收入",
		CreatedAt: now,
	})

	var wallet models.RiderWallet

	global.Db.Where("rider_id = ?", riderID).
		FirstOrCreate(&wallet, models.RiderWallet{
			RiderID:      riderID,
			Balance:      0,
			FrozenAmount: 0,
			TotalIncome:  0,
		})

	global.Db.Model(&wallet).Updates(map[string]interface{}{
		"balance":      gorm.Expr("balance + ?", order.TotalPrice),
		"total_income": gorm.Expr("total_income + ?", order.TotalPrice),
	})

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"success":   true,
			"actualFee": order.TotalPrice,
		},
	})
}

// GET /rider/orders/history
func GetOrderHistory(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")
	pageInt, _ := strconv.Atoi(page)
	sizeInt, _ := strconv.Atoi(size)
	offset := (pageInt - 1) * sizeInt

	statusStr := c.Query("status") // 可选，3:待接单, 4:配送中, 5:已完成
	date := c.Query("date")        // 可选

	type HistoryItem struct {
		ID          uint       `json:"id"`
		Restaurant  string     `json:"restaurant"`
		Customer    string     `json:"customer"`
		Fee         float64    `json:"fee"`
		Status      int        `json:"status"`
		CompletedAt *time.Time `json:"completedAt"`
	}

	var orders []models.Order
	var total int64

	query := global.Db.Model(&models.Order{}).
		Where("rider_id = ?", riderID)

	if statusStr != "" {
		if st, err := strconv.Atoi(statusStr); err == nil {
			query = query.Where("status = ?", st)
		}
	}

	if date != "" {
		query = query.Where("DATE(finish_at) = ?", date)
	}

	query.Count(&total)
	query.Order("finish_at DESC").
		Offset(offset).
		Limit(sizeInt).
		Find(&orders)

	list := make([]HistoryItem, 0, len(orders))

	for _, o := range orders {
		var merchant models.Merchant
		global.Db.Where("id = ?", o.MerchantID).First(&merchant)

		var consignee models.Consignee
		global.Db.Where("id = ?", o.Consigneeid).First(&consignee)

		item := HistoryItem{
			ID:          o.ID,
			Restaurant:  merchant.ShopName,
			Customer:    consignee.Name,
			Fee:         o.TotalPrice,
			Status:      o.Status,
			CompletedAt: o.FinishAt,
		}
		list = append(list, item)
	}

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"items": list,
			"total": total,
		},
	})
}

// GET /rider/orders/pickup
func GetPickupOrders(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	type PickupOrderResp struct {
		ID            uint   `json:"id"`
		Restaurant    string `json:"restaurant"`
		PickupAddress string `json:"pickupAddress"`
		PickupCode    string `json:"pickupCode"`
		ShopPhone     string `json:"shopPhone"`
		RemainingTime int    `json:"remainingTime"`
	}

	var orders []models.Order
	if err := global.Db.
		Where("rider_id = ? AND status = ? AND pickup_at IS NULL", riderID, 4).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	list := make([]PickupOrderResp, 0, len(orders))

	for _, o := range orders {
		var merchant models.Merchant
		global.Db.Where("id = ?", o.MerchantID).First(&merchant)

		remaining := 0
		if !o.ExpectedTime.IsZero() {
			diff := int(time.Until(o.ExpectedTime).Minutes())
			if diff > 0 {
				remaining = diff
			}
		}

		list = append(list, PickupOrderResp{
			ID:            o.ID,
			Restaurant:    merchant.ShopName,
			PickupAddress: merchant.ShopLocation,
			PickupCode:    o.PickupCode,
			ShopPhone:     merchant.Phone,
			RemainingTime: remaining,
		})
	}

	c.JSON(200, gin.H{"code": 1, "data": list})
}

// GET /rider/orders/:orderId
func GetOrderDetailForRider(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID")

	// 1. 查询订单
	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	// 2. 用 order.Consigneeid 查 consignee 表
	var consignee models.Consignee
	if err := global.Db.Where("id = ?", order.Consigneeid).First(&consignee).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "收货人不存在"})
		return
	}

	// 3. 用 consignee.Addressid 查 address 表
	var addr models.Address
	if err := global.Db.Where("id = ?", consignee.Addressid).First(&addr).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "地址不存在"})
		return
	}
	var merchant models.Merchant
	global.Db.Where("id = ?", order.MerchantID).First(&merchant)

	// 简单造一个时间线，至少有下单时间和完成时间
	timeline := []gin.H{
		{
			"label": "用户下单",
			"time":  order.CreatedAt,
		},
	}
	if !order.AcceptedAt.IsZero() {
		timeline = append(timeline, gin.H{
			"label": "骑手接单",
			"time":  order.AcceptedAt,
		})
	}
	if !order.PickupAt.IsZero() {
		timeline = append(timeline, gin.H{
			"label": "已取餐",
			"time":  order.PickupAt,
		})
	}
	if !order.FinishAt.IsZero() {
		timeline = append(timeline, gin.H{
			"label": "已送达",
			"time":  order.FinishAt,
		})
	}

	// 4. 返回前端想要的数据结构
	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"id":     order.ID,
			"status": order.Status,
			"total":  order.TotalPrice,

			"customerInfo": gin.H{
				"name":  consignee.Name,
				"phone": consignee.Phone,
				"address": gin.H{
					"province": addr.Province,
					"city":     addr.City,
					"district": addr.District,
					"street":   addr.Street,
					"detail":   addr.Detail,
				},
			},

			"shopInfo": gin.H{
				"name":    merchant.ShopName,
				"address": merchant.ShopLocation,
				"phone":   merchant.Phone,
			},

			"items":    []interface{}{}, // 暂时空数组
			"timeline": timeline,
		},
	})

}

// GET /rider/income/today
func GetTodayIncome(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var total float64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5 AND DATE(updated_at) = CURDATE()", riderID).
		Select("SUM(total_price)").Scan(&total)

	var count int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5 AND DATE(updated_at) = CURDATE()", riderID).
		Count(&count)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"todayIncome": total,
			"todayOrders": count,
		},
	})
}

// GET /rider/income/summary
func GetIncomeSummary(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var total float64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5", riderID).
		Select("SUM(total_price)").Scan(&total)

	var count int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5", riderID).
		Count(&count)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"totalIncome":     total,
			"completedOrders": count,
		},
	})
}

// GET /rider/income/month
func GetMonthIncome(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	type Item struct {
		Date  string  `json:"date"`
		Money float64 `json:"money"`
	}

	var data []Item

	global.Db.Raw(`
        SELECT DATE(updated_at) AS date, SUM(total_price) AS money
        FROM orders
        WHERE rider_id = ? AND status = 5
        GROUP BY DATE(updated_at)
        ORDER BY date ASC
    `, riderID).Scan(&data)

	c.JSON(200, gin.H{
		"code": 1,
		"data": data,
	})
}

// GET /rider/dashboard
func GetRiderDashboard(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	// 今日收入
	var todayIncome float64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5 AND DATE(updated_at)=CURDATE()", riderID).
		Select("SUM(total_price)").Scan(&todayIncome)

	// 今日完成单数
	var todayOrders int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5 AND DATE(updated_at)=CURDATE()", riderID).
		Count(&todayOrders)

	// 配送中：status = 4 且未完成
	var delivering int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = ? AND finish_at IS NULL", riderID, 4).
		Count(&delivering)

	// 待取货：status = 4 且 pickup_at IS NULL
	var waitPickup int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = ? AND pickup_at IS NULL", riderID, 4).
		Count(&waitPickup)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"todayIncome": todayIncome,
			"todayOrders": todayOrders,
			"delivering":  delivering,
			"waitPickup":  waitPickup,
		},
	})
}

// POST /rider/location
func UpdateRiderLocation(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Address   string  `json:"address"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	err := global.Db.Model(&models.RiderProfile{}).
		Where("user_id = ?", riderID).
		Updates(map[string]interface{}{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"address":   req.Address,
		}).Error

	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "定位更新失败"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": gin.H{"success": true}})
}
func AcceptOrder(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID")

	var order models.Order
	if err := global.Db.Where("id = ?", orderId).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在"})
		return
	}

	// 只有 status = 3 且尚未分配骑手的订单可以被抢
	if order.Status != 3 || order.RiderID != 0 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单已被抢走"})
		return
	}

	now := time.Now()
	pickupCode := utils.GeneratePickupCode()

	global.Db.Model(&order).Updates(map[string]interface{}{
		"status":      4, // 配送中
		"rider_id":    riderID,
		"pickup_code": pickupCode,
		"accepted_at": &now,
	})

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"success":    true,
			"pickupCode": pickupCode,
		},
	})
}

func GetIncomeStats(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	var dailyIncome, weeklyIncome, monthlyIncome float64
	var completedOrders int64

	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, todayStart).
		Select("SUM(amount)").Scan(&dailyIncome)

	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, weekStart).
		Select("SUM(amount)").Scan(&weeklyIncome)

	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, monthStart).
		Select("SUM(amount)").Scan(&monthlyIncome)

	// 完成订单：status = 5
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5").
		Count(&completedOrders)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"dailyIncome":     dailyIncome,
			"weeklyIncome":    weeklyIncome,
			"monthlyIncome":   monthlyIncome,
			"completedOrders": completedOrders,
		},
	})
}

// GET /rider/income/history
func GetIncomeHistory(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	// 前端 index.ts 注释里写了可以带 startDate / endDate
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var records []models.RiderIncomeRecord
	var total int64

	query := global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ?", riderID)

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	query.Count(&total)

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&records).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	// 手动组装前端期望的字段：
	// { id, orderId, amount, type, time, remark }
	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		items = append(items, gin.H{
			"id":      r.ID,
			"orderId": r.OrderID,
			"amount":  r.Amount,
			"type":    r.Type,
			"time":    r.CreatedAt, // ⭐ 关键：映射为 time
			"remark":  r.Remark,
		})
	}

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"items": items,
			"total": total,
		},
	})
}

func GetWeeklyStats(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	start := time.Now().AddDate(0, 0, -7)

	// 收入
	var weekIncome float64
	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, start).
		Select("SUM(amount)").Scan(&weekIncome)

	// 完成订单：status = 5
	var weekOrders int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 5 AND finish_at >= ?", riderID, start).
		Count(&weekOrders)

	var profile models.RiderProfile
	global.Db.Where("user_id = ?", riderID).First(&profile)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"weekIncome":  weekIncome,
			"weekOrders":  weekOrders,
			"onlineHours": profile.OnlineHours,
			"avgRating":   profile.Rating,
		},
	})
}

func GetWalletInfo(c *gin.Context) {
	id := c.GetUint("baseUserID")

	var wallet models.RiderWallet
	global.Db.Where("rider_id = ?", id).FirstOrCreate(&wallet)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"balance":      wallet.Balance,
			"frozenAmount": wallet.FrozenAmount,
			"totalIncome":  wallet.TotalIncome,
		},
	})
}

func Withdraw(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var req struct {
		Amount  float64 `json:"amount"`
		Account string  `json:"account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	now := time.Now()

	record := models.RiderWithdraw{
		RiderID:   riderID,
		Amount:    req.Amount,
		Account:   req.Account,
		Status:    "pending",
		AppliedAt: now,
	}

	global.Db.Create(&record)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"success":    true,
			"withdrawId": strconv.FormatUint(uint64(record.ID), 10), // ⭐ 转成字符串
		},
	})
}

func GetWithdrawHistory(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var list []models.RiderWithdraw
	global.Db.Where("rider_id = ?", riderID).
		Order("applied_at DESC").
		Find(&list)

	c.JSON(200, gin.H{
		"code": 1,
		"data": list,
	})
}
func GetDeliveryRoute(c *gin.Context) {
	orderId := c.Param("orderId")

	var route models.DeliveryRoute
	if err := global.Db.Where("order_id = ?", orderId).First(&route).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "未找到路线"})
		return
	}

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"route":         json.RawMessage(route.RouteJSON),
			"distance":      route.Distance,
			"estimatedTime": route.ETA,
		},
	})
}
