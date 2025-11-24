package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
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

    // 获取收获信息
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
			"price":dish.Dish.Price,
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
	fmt.Println(order.PayInfo.Paymethod,order.PayInfo.CheckoutTime,order.PayInfo.Packamount)
    response := gin.H{
        "code": 1,
        "data": gin.H{
            "id":        "o" + strconv.Itoa(int(order.ID)),
            "number":    order.CreatedAt.Format("20060102") + fmt.Sprintf("%06d", order.ID),
            "amount":    order.TotalPrice, // 假设Order表有TotalAmount字段
            "status":    order.Status,
            "orderTime": order.CreatedAt.Format(time.RFC3339),
			"phone":     consignee.Phone,
			"expected_time":order.ExpectedTime,
            "orderDetailList": items,
			"remark":order.Notes,
            "consignee": consignee.Name,
            "address":   address.Province + " " + address.City + " " + address.District + " " + address.Street + " " + address.Detail,
            "delivery": gin.H{
                "courierId":    "r" + strconv.Itoa(int(rider.ID)), // 使用rider.ID
                "courierName":  rider.RealName,                         // 使用rider.Name
                "courierPhone": rider.Phone,                        // 使用rider.Phone
            },
			"payMethod":order.PayInfo.Paymethod,
			"checkoutTime":order.PayInfo.CheckoutTime,
			"packAmount":order.PayInfo.Packamount,
			"deliveryAmount":order.PayInfo.Deliveryamount,
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
