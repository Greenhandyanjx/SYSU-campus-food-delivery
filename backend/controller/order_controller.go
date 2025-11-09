package controller

import (
	"backend/global"
	"backend/models"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

func GetOrderPage(c *gin.Context) {
    pageStr := c.Query("page")
    sizeStr := c.Query("size")
    beginStr := c.Query("begin")
    endStr := c.Query("end")
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
        beginTime, err = time.Parse("2006-01-02T15:04:05Z07:00", beginStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid begin time format", "data": nil})
            return
        }
    }
    if endStr != "" {
        endTime, err = time.Parse("2006-01-02T15:04:05Z07:00", endStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid end time format", "data": nil})
            return
        }
    }
    var orders []models.Order
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
    // 查询订单列表
    result := query.Limit(size).Offset(offset).Find(&orders)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order page", "data": nil})
        return
    }
    // 查询总订单数
    global.Db.Model(&models.Order{}).Where("created_at >= ? AND created_at <= ?", beginTime, endTime).Count(&count)
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "items": orders,
            "total": count,
        },
    })
}


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
    result := global.Db.Preload("Items").First(&order, orderId)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order detail", "data": nil})
        return
    }
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": order,
    })
}

func OrderAccept(c *gin.Context) {
	type OrderAcceptRequest struct {
       OrderID   string `json:"orderId" binding:"required"`
       AcceptBy  string `json:"acceptBy" binding:"required"`
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
    if order.Status != "pending" {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "order is not in pending state", "data": nil})
        return
    }
    // 更新订单状态为 'accepted'
    order.Status = "accepted"
    updateResult := global.Db.Save(&order)
    if updateResult.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to update order status", "data": nil})
        return
    }
    // 触发配送流程（这里假设配送流程是一个简单的消息通知）
    triggerDeliveryProcess(order)
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{"success": true},
    })
}

func triggerDeliveryProcess(order models.Order) {
    // 这里可以实现具体的配送流程逻辑，例如发送消息到配送员
    fmt.Printf("Delivery process triggered for order ID: %d\n", order.OrderID)
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
    result := global.Db.First(&order, "order_id = ?", request.OrderID)
    if result.Error != nil {
        if result.Error == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "order not found", "data": nil})
            return
        }
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order", "data": nil})
        return
    }
    // 检查订单状态是否可以拒单
    if order.Status != "pending" {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "order is not in pending state", "data": nil})
        return
	}
    // 更新订单状态为 'rejected'
    order.Status = "rejected"
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
    // 这里可以实现具体的用户通知逻辑，例如发送消息或邮件
    fmt.Printf("Notifying user of order ID: %d with reason: %s\n", order.OrderID, reason)
    // 实际应用中可能需要调用其他服务或发送消息
}