package controller

import (
	"backend/global"
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

//根据status查询order
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

//获取order列表，时间划分
//定义新的结构体，包含原有的 order 属性和 dishnames 字段
type OrderWithDishnames struct {
    ID     uint      `json:"order_id"`
    Dropofpoint   time.Time `json:"dropof_point"`
    Expected_time time.Time `json:"expected_time"`
    Phone       string    `json:"phone"`
    Status      int       `json:"status"`
    Numberoftableware    int       `json:"quantity"`
    TotalPrice  float64   `json:"total_price"`
    Dishnames   string    `json:"dishnames"`
    Notes string `json:"notes"`
    Consignee string `json:"consignee"` 
}
//赋值函数
func copyOrderFields(src *models.Order, dst *OrderWithDishnames) {
    dst.ID = src.ID
    dst.Dropofpoint = src.DropofPoint
    dst.Expected_time = src.ExpectedTime
    dst.Phone = src.Phone
    dst.Status = src.Status
    dst.Numberoftableware = src.Numberoftableware
    dst.TotalPrice = src.TotalPrice
    dst.Notes = src.Notes
    dst.Consignee = src.Consignee
}
func GetOrderPage(c *gin.Context) {
    pageStr := c.Query("page")
    sizeStr := c.Query("size")
    beginStr := c.Query("beginTime")
    endStr := c.Query("endTime")
    phonestr:= c.Query("phone")
    numberstr:= c.Query("number")
    status:=c.Query("status")
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1
    }
    size, err := strconv.Atoi(sizeStr)
    if err != nil || size < 1 {
        size = 20
    }
  
    // 解析时间范围
    var beginTime, endTime time.Time
    if beginStr != "" {
        beginTime, err = time.Parse("2006-01-02 15:04:05", beginStr)
        if err != nil {
            fmt.Println(err)
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid begin time format", "data": nil})
            return
        }
    }
    if endStr != "" {
        endTime, err = time.Parse("2006-01-02 15:04:05", endStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid end time format", "data": nil})
            return
        }
    }
    
    var orders []OrderWithDishnames
    var Orders []models.Order
    var orderIDs []uint
    var count int64
    // 计算分页偏移量
    offset := (page - 1) * size
    // 构建查询条件
    query := global.Db.Model(&models.Order{})
    if !beginTime.IsZero() {
        query = query.Where("created_at >= ?", beginTime)
    }
    if !endTime.IsZero() {
        query = query.Where("created_at <= ?", endTime)
    }
     if numberstr != "" {
        num, err := strconv.Atoi(numberstr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid number format", "data": nil})
            return
        }
        query = query.Where("ID= ?", num)
    }

    if phonestr != "" {
        query = query.Where("phone = ?", phonestr)
    }

    if status != "" {
        stat, err := strconv.Atoi(status)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid status format", "data": nil})
            return
        }
        query = query.Where("status = ?", stat)
    }

    // 查询订单列表

result := query.Limit(size).Offset(offset).Find(&Orders)
if result.Error != nil {
    c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order page", "data": nil})
    return
}
// 将查询结果复制到新的结构体切片
for _, srcOrder := range Orders {
        var dstOrder OrderWithDishnames
        copyOrderFields(&srcOrder, &dstOrder)
        orders = append(orders, dstOrder)
}

// 提取订单ID列表

for _, order := range orders {
    orderIDs = append(orderIDs, order.ID)
}
// 查询 orderdish 表以获取每个订单的菜品名
var orderDishnames []struct {
    OrderID   uint     `gorm:"column:order_id"`
    Dishnames string   `gorm:"column:dishnames"` // 修改为 string 类型，因为 GROUP_CONCAT 返回的是一个字符串
}
if err := global.Db.Table("order-dish").
    Select("order_id, GROUP_CONCAT(dishname) as dishnames").
    Where("order_id IN ?", orderIDs).
    Group("order_id").
    Find(&orderDishnames).Error; err != nil {
    log.Printf("查询订单菜品名失败: %v", err)
    c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询订单菜品名失败", "data": nil})
    return
}
// 构建订单和菜品名的映射
dishnamesMap := make(map[uint]string)
for _, od := range orderDishnames {
    dishnamesMap[od.OrderID] = od.Dishnames
}
   // 将映射中的数据赋值给结构体表
        for i, order := range orders {
            if dishnames, exists := dishnamesMap[order.ID]; exists {
                orders[i].Dishnames = dishnames
            }
        }
    // 查询总订单数
	countQuery := global.Db.Model(&models.Order{})
	if !beginTime.IsZero() {
		countQuery = countQuery.Where("created_at >= ?", beginTime)
	}
	if !endTime.IsZero() {
		countQuery = countQuery.Where("created_at <= ?", endTime)
	}
	countQuery.Count(&count)

    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "items": orders,
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
    var order models.Order
    result := global.Db.First(&order, orderId)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order detail", "data": nil})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": order,
    })
}

func OrderAccept(c *gin.Context) {
	type OrderAcceptRequest struct {
       OrderID   int `json:"id" binding:"required"`
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
  
    if err:=global.Db.Model(&models.Order{}).Where("id=?",order.ID).Update("status",3).Error;err!= nil {
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
    if order.Status != 2{
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
       OrderID   int `json:"id" binding:"required"`
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
  
    if err:=global.Db.Model(&models.Order{}).Where("id=?",order.ID).Update("status",4).Error;err!= nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
        return
    }
    // 触发配送流程（这里假设配送流程是一个简单的消息通知）
    triggerDeliveryProcess(order)
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "msg":"success",
    })
}

func OrderComplete(c *gin.Context) {
	type OrderAcceptRequest struct {
       OrderID   int `json:"id" binding:"required"`
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
  
    if err:=global.Db.Model(&models.Order{}).Where("id=?",order.ID).Update("status",5).Error;err!= nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
        return
    }
    // 触发后续流程（这里假设后续流程是一个简单的消息通知）
    triggerDeliveryProcess(order)
    //修改销量表
    // 查找对应的 dishId和num
    var orderDishes []models.OrderDish
    if err := global.Db.Model(&models.Order{}).Where("orderid = ?",request.OrderID).Find(&orderDishes).Error; err != nil {
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
        "msg":"success",
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