package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GET /rider/info
func GetRiderInfo(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID") // 从 AuthMiddleware 取用户ID

	var profile models.RiderProfile
	err := global.Db.Where("user_id = ?", baseUserID).First(&profile).Error
	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "未找到骑手信息"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": profile})
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

// 获取待接单订单（status=1）
func GetNewOrders(c *gin.Context) {
	type Result struct {
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

	var orders []models.Order
	err := global.Db.Where("status = ?", 1).
		Order("created_at DESC").
		Limit(20).
		Find(&orders).Error

	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	results := []Result{}

	for _, o := range orders {
		var merchant models.Merchant
		err := global.Db.Where("id = ?", o.MerchantID).First(&merchant).Error

		// 如果查不到商家，不要直接空白，给默认值
		if err != nil {
			merchant.ShopName = "未知商家"
			merchant.ShopLocation = "无地址"
		}

		result := Result{
			ID:              o.ID,
			Restaurant:      merchant.ShopName,
			PickupAddress:   merchant.ShopLocation,
			Customer:        "匿名用户",
			DeliveryAddress: o.Address,
			Distance:        1.2,
			EstimatedFee:    o.TotalPrice,
			EstimatedTime:   20,
			CreatedAt:       o.CreatedAt,
		}

		results = append(results, result)
	}

	c.JSON(200, gin.H{"code": 1, "data": results})
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

	if order.Status != 1 {
		tx.Rollback()
		c.JSON(200, gin.H{"code": 0, "msg": "订单已被抢走"})
		return
	}

	pickupCode := utils.GeneratePickupCode()

	err := tx.Model(&order).Updates(map[string]interface{}{
		"status":      2,
		"rider_id":    riderID,
		"pickup_code": pickupCode,
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
	if err := global.Db.Where("id = ?", orderId).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在"})
		return
	}

	// 骑手只能处理自己的订单
	if order.RiderID != riderID {
		c.JSON(200, gin.H{"code": 0, "msg": "无权限操作此订单"})
		return
	}

	// 状态必须是待取货（2）
	if order.Status != 2 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单状态不正确"})
		return
	}

	// 更新为配送中（3）
	err := global.Db.Model(&order).Updates(map[string]interface{}{
		"status":       3,
		"pickup_point": time.Now(),
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

	var orders []models.Order
	err := global.Db.Where("rider_id = ? AND status = 3", riderID).Find(&orders).Error
	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": orders})
}

// POST /rider/orders/:orderId/complete
func CompleteOrder(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	orderId := c.Param("orderId")

	var order models.Order
	err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error
	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	if order.Status != 3 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单状态不正确"})
		return
	}

	// 修改状态
	order.Status = 4
	order.DropofPoint = time.Now()

	if err := global.Db.Save(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "完成配送失败"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": gin.H{"success": true}})
}

// GET /rider/orders/history
func GetOrderHistory(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	// 分页参数
	page := c.DefaultQuery("page", "1")
	size := c.DefaultQuery("size", "10")

	pageInt, _ := strconv.Atoi(page)
	sizeInt, _ := strconv.Atoi(size)

	offset := (pageInt - 1) * sizeInt

	var orders []models.Order
	var total int64

	// 查询：当前骑手 + 已完成（status=4）
	query := global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4", riderID)

	// 统计数量
	query.Count(&total)

	// 分页查询
	query.Order("updated_at DESC").
		Offset(offset).
		Limit(sizeInt).
		Find(&orders)

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"items": orders,
			"total": total,
		},
	})
}

// GET /rider/orders/pickup
func GetPickupOrders(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var orders []models.Order
	err := global.Db.Where("rider_id = ? AND status = 2", riderID).Find(&orders).Error
	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": orders})
}

// GET /rider/orders/:orderId
func GetOrderDetailForRider(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID")

	// 1. 查订单
	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	// 2. 查收货人信息（Consignee）
	var consignee models.Consignee
	if err := global.Db.First(&consignee, order.Consigneeid).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "收货人信息查询失败"})
		return
	}

	// 3. 查地址（Address）
	var address models.Address
	global.Db.First(&address, consignee.Addressid)

	// 4. 查订单商品（如果你的表有 OrderDish）
	var dishes []models.OrderDish
	global.Db.Where("order_id = ?", order.ID).Find(&dishes)

	// 5. 封装返回（前端要求的结构）
	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"id":         order.ID,
			"status":     order.Status,
			"total":      order.TotalPrice,
			"pickupCode": order.PickupCode,

			"customerInfo": gin.H{
				"name":  consignee.Name,
				"phone": consignee.Phone,
				"address": gin.H{
					"province": address.Province,
					"city":     address.City,
					"district": address.District,
					"street":   address.Street,
					"detail":   address.Detail,
				},
			},

			"items": dishes,
		},
	})
}

// GET /rider/income/today
func GetTodayIncome(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var total float64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND DATE(updated_at) = CURDATE()", riderID).
		Select("SUM(total_price)").Scan(&total)

	var count int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND DATE(updated_at) = CURDATE()", riderID).
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
		Where("rider_id = ? AND status = 4", riderID).
		Select("SUM(total_price)").Scan(&total)

	var count int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4", riderID).
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
        WHERE rider_id = ? AND status = 4
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
		Where("rider_id = ? AND status = 4 AND DATE(updated_at)=CURDATE()", riderID).
		Select("SUM(total_price)").Scan(&todayIncome)

	// 今日单数
	var todayOrders int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND DATE(updated_at)=CURDATE()", riderID).
		Count(&todayOrders)

	// 配送中
	var delivering int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 3", riderID).
		Count(&delivering)

	// 待取货
	var waitPickup int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 2", riderID).
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
