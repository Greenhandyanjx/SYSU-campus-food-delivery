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
			// DeliveryAddress: o.PayInfo,
			Distance:        1.2,
			EstimatedFee:    o.TotalPrice,
			EstimatedTime:   20,
			CreatedAt:       o.CreatedAt,
		}

		results = append(results, result)
	}

	c.JSON(200, gin.H{"code": 1, "data": results})
}

// POST /rider/orders/:orderId/accept
func AcceptOrder(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID") // 骑手 ID，从 token 获取

	var order models.Order
	if err := global.Db.Where("id = ?", orderId).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在"})
		return
	}

	// 只能抢待接单订单
	if order.Status != 1 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单已被抢走"})
		return
	}

	// 生成取货码（例如 A123）
	pickupCode := utils.GeneratePickupCode()

	// 更新订单：status=2, rider_id, pickup_code
	updateErr := global.Db.Model(&order).Updates(map[string]interface{}{
		"status":      2,
		"rider_id":    riderID,
		"pickup_code": pickupCode,
	}).Error

	if updateErr != nil {
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "接单失败",
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 1,
		"data": gin.H{
			"success":    true,
			"pickupCode": pickupCode,
		},
	})
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
