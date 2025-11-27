package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
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
	// 1. 先查出待接单的订单
	var orders []models.Order
	if err := global.Db.
		Where("status = ?", 1).
		Order("created_at DESC").
		Limit(50).
		Find(&orders).Error; err != nil {

		log.Printf("GetNewOrders find orders error: %v\n", err)
		c.JSON(200, gin.H{
			"code": 0,
			"msg":  "获取订单失败",
			"data": []OrderItemResp{},
		})
		return
	}

	if len(orders) == 0 {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "获取成功",
			"data": []OrderItemResp{},
		})
		return
	}

	// 2. 收集所有需要的商家ID、收货人ID（去重）
	merchantIDs := make([]uint, 0)
	consigneeIDs := make([]uint, 0)
	merchantSet := make(map[uint]struct{})
	consigneeSet := make(map[uint]struct{})

	for _, o := range orders {
		// MerchantID 本身就是 uint
		if o.MerchantID > 0 {
			if _, ok := merchantSet[o.MerchantID]; !ok {
				merchantSet[o.MerchantID] = struct{}{}
				merchantIDs = append(merchantIDs, o.MerchantID)
			}
		}
		// Consigneeid 是 int，这里统一转成 uint
		if o.Consigneeid > 0 {
			cid := uint(o.Consigneeid)
			if _, ok := consigneeSet[cid]; !ok {
				consigneeSet[cid] = struct{}{}
				consigneeIDs = append(consigneeIDs, cid)
			}
		}
	}

	// 3. 一次性查出所有商家，map[uint]Merchant
	merchantMap := make(map[uint]models.Merchant)
	if len(merchantIDs) > 0 {
		var merchants []models.Merchant
		if err := global.Db.
			Where("id IN ?", merchantIDs).
			Find(&merchants).Error; err != nil {

			log.Printf("GetNewOrders find merchants error: %v\n", err)
		}
		for _, m := range merchants {
			merchantMap[m.ID] = m
		}
	}

	// 4. 一次性查出所有收货人，同时收集地址ID（地址这边用 int）
	consigneeMap := make(map[uint]models.Consignee)
	addressIDs := make([]int, 0)
	addrSet := make(map[int]struct{})

	if len(consigneeIDs) > 0 {
		var consignees []models.Consignee
		if err := global.Db.
			Where("id IN ?", consigneeIDs).
			Find(&consignees).Error; err != nil {

			log.Printf("GetNewOrders find consignees error: %v\n", err)
		}
		for _, cg := range consignees {
			consigneeMap[cg.ID] = cg

			if cg.Addressid > 0 {
				aid := cg.Addressid // int
				if _, ok := addrSet[aid]; !ok {
					addrSet[aid] = struct{}{}
					addressIDs = append(addressIDs, aid)
				}
			}
		}
	}

	// 5. 一次性查出所有地址，map[int]Address
	addrMap := make(map[int]models.Address)
	if len(addressIDs) > 0 {
		var addrs []models.Address
		if err := global.Db.
			Where("id IN ?", addressIDs).
			Find(&addrs).Error; err != nil {

			log.Printf("GetNewOrders find addresses error: %v\n", err)
		}
		for _, a := range addrs {
			addrMap[a.ID] = a // a.ID 是 int，这里就不会再报错了
		}
	}

	// 6. 组装返回数据，遇到脏数据就跳过
	list := make([]OrderItemResp, 0, len(orders))

	for _, o := range orders {
		// 商家
		m, ok := merchantMap[o.MerchantID]
		if !ok {
			log.Printf("order %d merchant %d not found\n", o.ID, o.MerchantID)
			continue
		}

		// 收货人
		cid := uint(o.Consigneeid)
		cg, ok := consigneeMap[cid]
		if !ok {
			log.Printf("order %d consignee %d not found\n", o.ID, o.Consigneeid)
			continue
		}

		// 地址（全部用 int）
		aid := cg.Addressid
		addr, ok := addrMap[aid]
		if !ok {
			log.Printf("order %d address %d not found\n", o.ID, cg.Addressid)
			continue
		}

		fullAddr := addr.Province + addr.City + addr.District + addr.Street + addr.Detail

		list = append(list, OrderItemResp{
			ID:              o.ID,
			Restaurant:      m.ShopName,
			PickupAddress:   m.ShopLocation,
			Customer:        cg.Name,
			DeliveryAddress: fullAddr,
			Distance:        1.2,
			EstimatedFee:    o.TotalPrice,
			EstimatedTime:   20,
			CreatedAt:       o.CreatedAt,
		})
	}

	log.Printf("GetNewOrders success, count=%d\n", len(list))

	c.JSON(200, gin.H{
		"code": 1,
		"msg":  "获取成功",
		"data": list,
	})
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
	now := time.Now()
	err := global.Db.Model(&order).Updates(map[string]interface{}{
		"status":    3,
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
		Where("rider_id = ? AND status = 3", riderID).
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
			CustomerAvatar:  "", // 先不给头像字段，避免 bu.Avatar 报错
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

	if order.Status != 3 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单状态不正确"})
		return
	}

	now := time.Now()

	// ==== 1. 正确更新订单（不用 Save） ====
	if err := global.Db.Model(&models.Order{}).
		Where("id = ?", order.ID).
		Updates(map[string]interface{}{
			"status":       4,
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

	// 必须提供 default 值才能创建新记录
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
			"actualFee": order.TotalPrice, // 实际配送费
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

	statusStr := c.Query("status") // 可选，根据 index.ts 说明
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
		// 前端用字符串的话你可以自己映射一下
		// 例如 "completed" -> 4
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
			CompletedAt: o.FinishAt, // 记得在 models.Order 里有 FinishAt
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
		Where("rider_id = ? AND status = 2", riderID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	list := make([]PickupOrderResp, 0, len(orders))

	for _, o := range orders {
		var merchant models.Merchant
		global.Db.Where("id = ?", o.MerchantID).First(&merchant)

		// 剩余时间：用期望送达时间减当前时间，单位分钟，负数就置 0
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
			ShopPhone:     merchant.Phone, // 根据你 Merchant 实际字段改一下
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
func AcceptOrder(c *gin.Context) {
	orderId := c.Param("orderId")
	riderID := c.GetUint("baseUserID")

	var order models.Order
	if err := global.Db.Where("id = ?", orderId).First(&order).Error; err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "订单不存在"})
		return
	}

	if order.Status != 1 {
		c.JSON(200, gin.H{"code": 0, "msg": "订单已被抢走"})
		return
	}

	now := time.Now()
	pickupCode := utils.GeneratePickupCode()

	global.Db.Model(&order).Updates(map[string]interface{}{
		"status":      2,
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
	// period := c.DefaultQuery("period", "today") // 如不需要，可以忽略

	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	weekStart := now.AddDate(0, 0, -7)
	monthStart := now.AddDate(0, -1, 0)

	var dailyIncome, weeklyIncome, monthlyIncome float64
	var completedOrders int64

	// 今天收入
	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, todayStart).
		Select("SUM(amount)").Scan(&dailyIncome)

	// 七天收入
	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, weekStart).
		Select("SUM(amount)").Scan(&weeklyIncome)

	// 一个月收入
	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, monthStart).
		Select("SUM(amount)").Scan(&monthlyIncome)

	// 总完成单数（你可以按需求改成最近一月）
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4").
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

	// 完成订单
	var weekOrders int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Count(&weekOrders)

	// 在线时长（你 RiderProfile 里应有 OnlineHours 字段）
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
