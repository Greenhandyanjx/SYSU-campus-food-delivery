package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// 根据status查询order
func GetOrderListByStatus(c *gin.Context) {
	status := c.Query("status")
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 20
	}
	var orders []models.Order
	var count int64
	// 计算分页偏移量
	offset := (page - 1) * size
	// 如果请求 status==0（pending），为了避免商家/骑手看到用户未完成的 pending 订单，直接返回空
	if status == "0" {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"items": []models.Order{}, "total": 0}})
		return
	}
	// 查询订单列表
	result := global.Db.Model(&models.Order{}).Where("status = ?", status).Limit(size).Offset(offset).Find(&orders)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order list", "data": nil})
		return
	}
	// 查询总订单数
	global.Db.Model(&models.Order{}).Where("status = ?", status).Count(&count)
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items": orders,
			"total": count,
		},
	})
}

// 获取order列表，时间划分
func GetOrderPage(c *gin.Context) {
	pageStr := c.Query("page")
	sizeStr := c.Query("size")
	beginStr := c.Query("beginTime")
	endStr := c.Query("endTime")
	phonestr := c.Query("phone")
	numberstr := c.Query("number")
	status := c.Query("status")
	page, size, beginTime, endTime := utils.ParsePaginationAndTime(c, pageStr, sizeStr, beginStr, endStr)
	if page == 0 || size == 0 {
		return
	}
	orders, count, err := utils.FetchOrders(c, page, size, beginTime, endTime, phonestr, numberstr, status)
	if err != nil {
		return
	}
	consigneeMap, addressMap := utils.FetchConsigneesAndAddresses(c, orders)
	ordersWithDetails := utils.CopyOrdersToOrderWithDishnames(orders, consigneeMap, addressMap)
	utils.FetchDishnames(c, &ordersWithDetails)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items": ordersWithDetails,
			"total": count,
		},
	})

}

// 根据orderId获取订单详情
func GetOrderDetail(c *gin.Context) {
	orderIdStr := c.Query("orderId")
	if orderIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "orderId is required", "data": nil})
		return
	}
	orderId, err := strconv.Atoi(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid orderId format", "data": nil})
		return
	}
	// 获取订单基础信息
	var order models.Order
	result := global.Db.Preload("PayInfo").First(&order, orderId)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order detail", "data": nil})
		return
	}
	// 获取收货信息
	var consignee models.Consignee
	result = global.Db.First(&consignee, "id = ?", order.Consigneeid)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get consignee detail", "data": nil})
		return
	}

	// 获取收获地址信息
	var address models.Address
	result = global.Db.First(&address, consignee.Addressid)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get address detail", "data": nil})
		return
	}

	// 获取订单中的餐品信息并关联Meal表
	var orderMeals []models.OrderMeal
	result = global.Db.Preload("Meal").Table("order_meals").Where("order_id = ?", orderId).Find(&orderMeals)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order meals", "data": nil})
		return
	}
	// 获取订单中的菜品信息
	var orderDishes []models.OrderDish
	result = global.Db.Preload("Dish").Table("order_dishes").Where("order_id = ?", orderId).Find(&orderDishes)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order dishes", "data": nil})
		return
	}
	// 构建items信息
	items := make([]gin.H, 0)
	for _, orderMeal := range orderMeals {
		meal := orderMeal.Meal // 假设OrderMeal表中预加载了Meal信息
		items = append(items, gin.H{
			"skuId": "m" + strconv.Itoa(orderMeal.MealID),
			"name":  meal.Mealname,
			"qty":   orderMeal.Num,
			"price": meal.Price, // 也可以根据需要添加价格信息
		})
	}
	for _, dish := range orderDishes {
		items = append(items, gin.H{
			"skuId": "d" + strconv.Itoa(dish.DishID),
			"name":  dish.Dish.DishName,
			"qty":   dish.Num,
			"price": dish.Dish.Price,
		})
	}
	// 获取配送员信息
	var rider models.Rider
	result = global.Db.First(&rider, order.RiderID)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get rider detail", "data": nil})
		return
	}
	// 构建最终返回的数据
	fmt.Println(order.PayInfo.Paymethod, order.PayInfo.CheckoutTime, order.PayInfo.Packamount)
	response := gin.H{
		"code": 1,
		"data": gin.H{
			"id":              "o" + strconv.Itoa(int(order.ID)),
			"number":          order.CreatedAt.Format("20060102") + fmt.Sprintf("%06d", order.ID),
			"amount":          order.TotalPrice, // 假设Order表有TotalAmount字段
			"status":          order.Status,
			"orderTime":       order.CreatedAt.Format(time.RFC3339),
			"phone":           consignee.Phone,
			"expected_time":   order.ExpectedTime,
			"orderDetailList": items,
			"remark":          order.Notes,
			"consignee":       consignee.Name,
			"address":         address.Province + " " + address.City + " " + address.District + " " + address.Street + " " + address.Detail,
			"delivery": gin.H{
				"courierId":    "r" + strconv.Itoa(int(rider.ID)), // 使用rider.ID
				"courierName":  rider.RealName,                    // 使用rider.Name
				"courierPhone": rider.Phone,                       // 使用rider.Phone
			},
			"payMethod":      order.PayInfo.Paymethod,
			"checkoutTime":   order.PayInfo.CheckoutTime,
			"packAmount":     order.PayInfo.Packamount,
			"deliveryAmount": order.PayInfo.Deliveryamount,
		},
	}
	c.JSON(http.StatusOK, response)
}

func OrderAccept(c *gin.Context) {
	type OrderAcceptRequest struct {
		OrderID int `json:"id" binding:"required"`
	}
	var request OrderAcceptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid request body", "data": nil})
		return
	}
	var order models.Order
	result := global.Db.First(&order, request.OrderID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order", "data": nil})
		return
	}
	// 检查订单状态是否可以接受
	if order.Status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "order is not in pending state", "data": nil})
		return
	}
	// 更新订单状态为 'accepted'
	order.Status = 3

	if err := global.Db.Model(&models.Order{}).Where("id=?", order.ID).Update("status", 3).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
		return
	}
	// 触发配送流程（这里假设配送流程是一个简单的消息通知）
	triggerDeliveryProcess(order)
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{"success": true},
	})
}

func triggerDeliveryProcess(order models.Order) {
	fmt.Printf("Delivery process triggered for order ID: %d\n", order.ID)
	// 实际应用中可能需要调用其他服务或发送消息
}

func OrderReject(c *gin.Context) {
	type OrderRejectRequest struct {
		OrderID string `json:"orderId" binding:"required"`
		Reason  string `json:"reason" binding:"required"`
	}

	var request OrderRejectRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid request body", "data": nil})
		return
	}
	var order models.Order
	result := global.Db.First(&order, "ID = ?", request.OrderID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order", "data": nil})
		return
	}
	// 检查订单状态是否可以拒单
	if order.Status != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "order is not in pending state", "data": nil})
		return
	}
	// 更新订单状态为 'rejected'
	order.Status = 6
	updateResult := global.Db.Save(&order)
	if updateResult.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
		return
	}
	// 通知用户（这里假设通知用户是一个简单的消息通知）
	notifyUser(order, request.Reason)
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{"success": true},
	})
}

func notifyUser(order models.Order, reason string) {
	fmt.Printf("Notifying user of order ID: %d with reason: %s\n", order.ID, reason)
	// 实际应用中可能需要调用其他服务或发送消息
}

func OrderDelivery(c *gin.Context) {
	type OrderAcceptRequest struct {
		OrderID int `json:"id" binding:"required"`
	}
	var request OrderAcceptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid request body", "data": nil})
		return
	}
	var order models.Order
	result := global.Db.First(&order, request.OrderID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order", "data": nil})
		return
	}
	// 检查订单状态是否可以接受
	if order.Status != 3 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "order is not in right state", "data": nil})
		return
	}

	if err := global.Db.Model(&models.Order{}).Where("id=?", order.ID).Update("status", 4).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
		return
	}
	// 触发配送流程（这里假设配送流程是一个简单的消息通知）
	triggerDeliveryProcess(order)
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "success",
	})
}

func OrderComplete(c *gin.Context) {
	type OrderAcceptRequest struct {
		OrderID int `json:"id" binding:"required"`
	}
	var request OrderAcceptRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid request body", "data": nil})
		return
	}
	var order models.Order
	result := global.Db.First(&order, request.OrderID)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order", "data": nil})
		return
	}
	// 检查订单状态是否可以接受
	if order.Status != 4 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "order is not in right state", "data": nil})
		return
	}

	if err := global.Db.Model(&models.Order{}).Where("id=?", order.ID).Update("status", 5).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
		return
	}
	// 触发后续流程（这里假设后续流程是一个简单的消息通知）
	triggerDeliveryProcess(order)
	//修改销量表
	// 查找对应的 dishId和num
	var orderDishes []models.OrderDish
	if err := global.Db.Model(&models.Order{}).Where("orderid = ?", request.OrderID).Find(&orderDishes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order dishes", "data": nil})
		return
	}

	// 更新 sales_stats 表中的 quantity
	for _, od := range orderDishes {
		if err := global.Db.Model(&models.SalesStat{}).
			Where("merchant_id = ? AND item_type = ? AND item_id = ? AND date = ?",
				order.MerchantID, "dish", od.DishID, order.CreatedAt.Format("2006-01-02")).
			Updates(map[string]interface{}{"quantity": gorm.Expr("quantity + ?", od.Num)}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update sales stats", "data": nil})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "message": "sales stats updated successfully"})
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"msg":  "success",
	})
}

func Orderadd(c *gin.Context) {
	var newOrder models.Order
	// 绑定请求体到 Order 结构体
	if err := c.ShouldBindJSON(&newOrder); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// 设置默认值或其他逻辑处理
	if newOrder.PickupPoint.IsZero() {
		newOrder.PickupPoint = time.Now()
	}
	if newOrder.DropofPoint.IsZero() {
		newOrder.DropofPoint = time.Now()
	}
	if newOrder.ExpectedTime.IsZero() {
		newOrder.ExpectedTime = time.Now()
	}
	// 创建订单记录
	result := global.Db.Table("orders").Create(&newOrder)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}
	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "order added successfully", "order": newOrder})
}

// CreatePayOrder 创建一个支付订单（预下单），兼容单商家和批量 shops 格式
func CreatePayOrder(c *gin.Context) {
	type ShopReq struct {
		MerchantID uint    `json:"merchantId" binding:"required"`
		TotalPrice float64 `json:"totalPrice" binding:"required"`
	}
	type Req struct {
		Shops       []ShopReq `json:"shops"`
		MerchantID  uint      `json:"merchantId"` // 兼容旧字段
		Consigneeid int       `json:"consigneeid"`
		TotalPrice  float64   `json:"totalPrice"`
		Remarks     string    `json:"remarks"`
	}

	var req Req
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid request", "error": err.Error()})
		return
	}

	// 获取用户 ID（中间件写入的 baseUserID），并校验 consignee 属于该用户
	baseUserIDIface, exists := c.Get("baseUserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "message": "not authenticated"})
		return
	}
	baseUserID := baseUserIDIface.(uint)

	// 如果没有传 consigneeid，尝试查找当前用户的第一个 consignee
	var consignee models.Consignee
	if req.Consigneeid == 0 {
		if err := global.Db.Where("userid = ?", baseUserID).First(&consignee).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "no consignee found for user, please provide consigneeid"})
			return
		}
		req.Consigneeid = int(consignee.ID)
	} else {
		if err := global.Db.First(&consignee, req.Consigneeid).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid consignee"})
			return
		}
		if consignee.Userid != baseUserID {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "message": "consignee does not belong to user"})
			return
		}
	}

	// 兼容：如果没有提供 shops 数组，但提供了单个 merchantId+totalPrice，转为单元素 shops
	if len(req.Shops) == 0 && req.MerchantID != 0 && req.TotalPrice > 0 {
		req.Shops = []ShopReq{{MerchantID: req.MerchantID, TotalPrice: req.TotalPrice}}
	}

	if len(req.Shops) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "no shops provided"})
		return
	}

	// 统一为所有商家创建单个 payinfo（总价），并为每个商家创建 order 关联同一个 payinfo
	// 这样前端只需要展示一个二维码（总金额）
	type RespItem struct {
		OrderID    uint   `json:"orderId"`
		OutTradeNo string `json:"out_trade_no"`
		CodeURL    string `json:"code_url"`
		MerchantID uint   `json:"merchantId"`
	}
	var resp []RespItem
	// 计算总价
	var totalSum float64 = 0
	for _, s := range req.Shops {
		totalSum += s.TotalPrice
	}

	// 生成单个 out_trade_no/code_url
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to generate out_trade_no"})
		return
	}
	outTradeNo := fmt.Sprintf("o_%d_%s", time.Now().Unix(), hex.EncodeToString(b))
	codeURL := fmt.Sprintf("https://pay.example.local/qr?out_trade_no=%s", outTradeNo)

	// 创建单个 payinfo
	tx := global.Db.Begin()
	pay := models.PayInfo{
		Paymethod:      1,
		Packamount:     0,
		CheckoutTime:   time.Now(),
		Deliveryamount: 0,
		OutTradeNo:     outTradeNo,
		CodeURL:        codeURL,
		Status:         "pending",
		// 可扩展字段记录总价
		// TotalAmount:    totalSum, // 如果模型有该字段
	}
	exp := time.Now().Add(15 * time.Minute)
	pay.ExpiresAt = &exp

	if err := tx.Create(&pay).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to create payinfo"})
		return
	}

	// 为每个商家创建或升级 order，关联同一个 payinfo
	// 先查找当前用户是否已有 pending (status=0) 订单，若存在并匹配 merchant，则升级该订单为待支付状态并关联本次 payinfo
	var pendingOrders []models.Order
	if err := tx.Where("status = ? AND consigneeid = ?", 0, req.Consigneeid).Find(&pendingOrders).Error; err == nil {
		// pendingOrders loaded
	}

	// 构建一个 map 以便快速匹配商家
	pendingMap := make(map[uint]*models.Order)
	for i := range pendingOrders {
		o := &pendingOrders[i]
		pendingMap[o.MerchantID] = o
	}

	for _, s := range req.Shops {
		if po, ok := pendingMap[s.MerchantID]; ok {
			// 升级现有 pending order
			if err := tx.Model(&models.Order{}).Where("id = ?", po.ID).Updates(map[string]interface{}{"status": 1, "pay_infoid": int(pay.ID), "total_price": s.TotalPrice}).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to upgrade pending order"})
				return
			}
			resp = append(resp, RespItem{OrderID: uint(po.ID), OutTradeNo: outTradeNo, CodeURL: codeURL, MerchantID: s.MerchantID})
			continue
		}

		// 无 pending order，创建新 order
		order := models.Order{
			Consigneeid:  req.Consigneeid,
			PickupPoint:  time.Now(),
			DropofPoint:  time.Now(),
			ExpectedTime: time.Now(),
			Status:       1, // 1 = unpaid/created
			TotalPrice:   s.TotalPrice,
			MerchantID:   s.MerchantID,
			Notes:        req.Remarks,
			PayInfoid:    int(pay.ID),
			Userid:       baseUserID,
		}
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to create order"})
			return
		}
		resp = append(resp, RespItem{OrderID: uint(order.ID), OutTradeNo: outTradeNo, CodeURL: codeURL, MerchantID: s.MerchantID})
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to commit transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"code_url": codeURL, "out_trade_no": outTradeNo, "orders": resp}})
}

// CreatePendingOrder 创建一个“预览/待支付”订单，用于用户进入结算页时持久化未完成的尝试。
func CreatePendingOrder(c *gin.Context) {
	type ShopReq struct {
		MerchantID uint    `json:"merchantId" binding:"required"`
		TotalPrice float64 `json:"totalPrice" binding:"required"`
	}
	type Req struct {
		Shops       []ShopReq `json:"shops"`
		MerchantID  uint      `json:"merchantId"` // 兼容旧字段
		Consigneeid int       `json:"consigneeid"`
		TotalPrice  float64   `json:"totalPrice"`
		Remarks     string    `json:"remarks"`
	}

	var req Req
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid request", "error": err.Error()})
		return
	}

	// 获取用户 ID（中间件写入的 baseUserID），并校验 consignee 属于该用户
	baseUserIDIface, exists := c.Get("baseUserID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "message": "not authenticated"})
		return
	}
	baseUserID := baseUserIDIface.(uint)

	// 如果没有传 consigneeid，尝试查找当前用户的第一个 consignee
	var consignee models.Consignee
	if req.Consigneeid == 0 {
		if err := global.Db.Where("userid = ?", baseUserID).First(&consignee).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "no consignee found for user, please provide consigneeid"})
			return
		}
		req.Consigneeid = int(consignee.ID)
	} else {
		if err := global.Db.First(&consignee, req.Consigneeid).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid consignee"})
			return
		}
		if consignee.Userid != baseUserID {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "message": "consignee does not belong to user"})
			return
		}
	}

	// 兼容：如果没有提供 shops 数组，但提供了单个 merchantId+totalPrice，转为单元素 shops
	if len(req.Shops) == 0 && req.MerchantID != 0 && req.TotalPrice > 0 {
		req.Shops = []ShopReq{{MerchantID: req.MerchantID, TotalPrice: req.TotalPrice}}
	}

	if len(req.Shops) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "no shops provided"})
		return
	}

	// 创建一个 payinfo 记录以便关联（schema 要求 payinfoid 非空）
	tx := global.Db.Begin()
	pay := models.PayInfo{
		Paymethod:      0,
		Packamount:     0,
		CheckoutTime:   time.Now(),
		Deliveryamount: 0,
		OutTradeNo:     "",
		CodeURL:        "",
		Status:         "pending",
	}
	exp := time.Now().Add(24 * time.Hour)
	pay.ExpiresAt = &exp
	if err := tx.Create(&pay).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to create payinfo"})
		return
	}

	var resp []map[string]interface{}
	for _, s := range req.Shops {
		order := models.Order{
			Consigneeid:  req.Consigneeid,
			PickupPoint:  time.Now(),
			DropofPoint:  time.Now(),
			ExpectedTime: time.Now(),
			Status:       0, // 0 = pending (user initiated, not visible to merchant/rider)
			TotalPrice:   s.TotalPrice,
			MerchantID:   s.MerchantID,
			Notes:        req.Remarks,
			PayInfoid:    int(pay.ID),
			Userid:       baseUserID,
		}
		if err := tx.Create(&order).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to create pending order"})
			return
		}
		// 将该用户购物车中对应商家的 items 迁移为 order_dishes（并删除购物车项），以便在 pending 状态下保留商品明细
		var cart models.Cart
		if err := tx.Where("user_id = ?", baseUserID).First(&cart).Error; err == nil {
			var items []models.CartItem
			if err := tx.Where("cart_id = ? AND merchant_id = ?", cart.ID, s.MerchantID).Find(&items).Error; err == nil {
				for _, it := range items {
					od := models.OrderDish{OrderID: int(order.ID), DishID: int(it.DishID), Num: it.Qty}
					if err := tx.Create(&od).Error; err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to create order dish"})
						return
					}
				}
				// 删除已迁移的购物车项，避免后续重复迁移
				if err := tx.Where("cart_id = ? AND merchant_id = ?", cart.ID, s.MerchantID).Delete(&models.CartItem{}).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to cleanup cart items after pending creation"})
					return
				}
			}
		}
		resp = append(resp, map[string]interface{}{"orderId": order.ID, "merchantId": s.MerchantID})
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to commit"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"orders": resp}})
}

// GetOrderStatus 返回订单支付及状态信息
func GetOrderStatus(c *gin.Context) {
	orderIdStr := c.Query("orderId")
	if orderIdStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "orderId is required"})
		return
	}
	oid, err := strconv.Atoi(orderIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid orderId"})
		return
	}
	var order models.Order
	if err := global.Db.Preload("PayInfo").First(&order, oid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"orderId": order.ID, "status": order.Status, "pay_status": order.PayInfo.Status}})
}

// PaymentNotify 支付平台回调处理（notify_url）
func PaymentNotify(c *gin.Context) {
	// 解析回调：支持 JSON 格式的简单回调 { out_trade_no: "...", result: "SUCCESS" }
	var payload map[string]interface{}
	if err := c.ShouldBindJSON(&payload); err != nil {
		// 也尝试表单形式
		if err := c.Request.ParseForm(); err != nil {
			c.String(http.StatusBadRequest, "invalid callback")
			return
		}
		payload = map[string]interface{}{}
		for k, v := range c.Request.PostForm {
			if len(v) > 0 {
				payload[k] = v[0]
			}
		}
	}
	out, _ := payload["out_trade_no"].(string)
	result, _ := payload["result"].(string)
	if out == "" {
		c.String(http.StatusBadRequest, "missing out_trade_no")
		return
	}

	var pay models.PayInfo
	if err := global.Db.Where("out_trade_no = ?", out).First(&pay).Error; err != nil {
		c.String(http.StatusNotFound, "payinfo not found")
		return
	}

	if result == "SUCCESS" {
		now := time.Now()
		pay.Status = "paid"
		pay.PaidAt = &now
		if err := global.Db.Save(&pay).Error; err != nil {
			c.String(http.StatusInternalServerError, "failed to update payinfo")
			return
		}

		// 找到所有与此 payinfo 关联的订单，并逐个将其状态设为已支付(2)，同时把购物车中对应商家的商品写入 order_dishes
		var orders []models.Order
		if err := global.Db.Where("pay_infoid = ?", pay.ID).Find(&orders).Error; err != nil {
			c.String(http.StatusOK, "ok")
			return
		}

		// 对每个子订单执行迁移（在同一个事务内逐个处理）
		for _, order := range orders {
			tx := global.Db.Begin()
			if err := tx.Model(&models.Order{}).Where("id = ?", order.ID).Update("status", 2).Error; err != nil {
				tx.Rollback()
				c.String(http.StatusInternalServerError, "failed to update order status")
				return
			}

			// 从 carts 找到当前用户的 cart 并迁移其 cart_items 到 order_dishes（按 merchant 分组迁移）
			var consignee models.Consignee
			if err := tx.First(&consignee, order.Consigneeid).Error; err == nil {
				var cart models.Cart
				if err := tx.Where("user_id = ?", consignee.Userid).First(&cart).Error; err == nil {
					var items []models.CartItem
					// 如果该订单已存在 order_dishes（可能在 pending 创建时已迁移），则跳过迁移以避免重复
					var existCount int64
					tx.Model(&models.OrderDish{}).Where("order_id = ?", order.ID).Count(&existCount)
					if existCount == 0 {
						if err := tx.Where("cart_id = ? AND merchant_id = ?", cart.ID, order.MerchantID).Find(&items).Error; err == nil {
							for _, it := range items {
								od := models.OrderDish{OrderID: int(order.ID), DishID: int(it.DishID), Num: it.Qty}
								if err := tx.Create(&od).Error; err != nil {
									tx.Rollback()
									c.String(http.StatusInternalServerError, "failed to create order dish")
									return
								}
							}
							// 删除已迁移的购物车项
							if err := tx.Where("cart_id = ? AND merchant_id = ?", cart.ID, order.MerchantID).Delete(&models.CartItem{}).Error; err != nil {
								tx.Rollback()
								c.String(http.StatusInternalServerError, "failed to cleanup cart items")
								return
							}
						}
					}
				}
			}

			// 分配骑手（简化：暂不实现真实分配，留空或记录日志）
			fmt.Printf("Order %d paid, merchant %d — assign rider later\n", order.ID, order.MerchantID)

			if err := tx.Commit().Error; err != nil {
				c.String(http.StatusInternalServerError, "tx commit failed")
				return
			}
		}
	}

	// 返回平台要求的响应
	c.String(http.StatusOK, "success")
}
