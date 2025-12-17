package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetBusinessData 获取店铺的经营概览
func GetBusinessData(c *gin.Context) {
    baseid, ok := c.Get("baseUserID")
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }
    merchantID, ok := baseid.(uint)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "商户ID类型错误"})
        return
    }

    currentDate := time.Now().Format("2006-01-02")
    cacheKey := fmt.Sprintf("business-data-overview:merchant_id:%d:date:%s", merchantID, currentDate)

    // 尝试从 Redis 获取缓存的数据
    var cachedData struct {
        Turnover             float64 `json:"turnover"`
        ValidOrderCount      int64   `json:"validOrderCount"`
        OrderCompletionRate  float64 `json:"orderCompletionRate"`
        UnitPrice            float64 `json:"unitPrice"`
        NewUsers             int64   `json:"newUsers"`
    }
    found, err := utils.GetJSON(context.Background(), cacheKey, &cachedData)
    if err != nil {
        log.Printf("Redis读取错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    if found {
        c.JSON(http.StatusOK, gin.H{
            "code": 1,
            "data": gin.H{
                "turnover": cachedData.Turnover,
                "validOrderCount": cachedData.ValidOrderCount,
                "orderCompletionRate": cachedData.OrderCompletionRate,
                "unitPrice": cachedData.UnitPrice,
                "newUsers": cachedData.NewUsers,
            },
        })
        return
    }

    var todaydata models.Revenue
    // 查询 revenue 表，获取对应日期的营业额
    if err := global.Db.Table("revenues").
        Where("merchant_id = ? AND day = ?", baseid, currentDate).
        First(&todaydata).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "数据未找到"})
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        }
        return
    }

    // 初始化计数变量
    var countBetween2And6, countForStatus5 int64

    // 查询状态为大于2小于6的订单数量
    if err := global.Db.Model(&models.Order{}).
        Where("status > ? AND status < ? AND merchant_id = ? AND DATE(pickup_point) = ?", 2, 6, baseid, currentDate).
        Count(&countBetween2And6).Error; err != nil {
        log.Printf("查询状态为大于2小于6的订单数量失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }

    // 查询状态为5的订单数量
    if err := global.Db.Model(&models.Order{}).
        Where("status = ? AND merchant_id = ? AND DATE(pickup_point) = ?", 5, baseid, currentDate).
        Count(&countForStatus5).Error; err != nil {
        log.Printf("查询状态为5的订单数量失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }

    totalRevenue := todaydata.Revenue
    // 计算平均价格
    var avgTicket float64
    var orderCompletionRate float64
    if countBetween2And6 > 0 {
        avgTicket = totalRevenue / float64(countBetween2And6)
        orderCompletionRate = float64(countForStatus5) / float64(countBetween2And6)
    } else {
        avgTicket = 0
        orderCompletionRate = 0
    }

    fmt.Println(totalRevenue, todaydata.Usernumber, avgTicket, orderCompletionRate, countForStatus5, countBetween2And6)

    // 准备返回数据
    data := gin.H{
        "turnover":              totalRevenue,
        "validOrderCount":       countBetween2And6,
        "orderCompletionRate":   orderCompletionRate,
        "unitPrice":             avgTicket,
        "newUsers":              todaydata.Usernumber,
    }

    // 序列化数据并存入 Redis
    err = utils.SetJSON(context.Background(), cacheKey, data, 5*time.Minute)
    if err != nil {
        log.Printf("Redis写入错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }

    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": data,
    })
}


// GetOrderData 获取当日订单统计
func GetOrderData(c *gin.Context) {
    merchantIDInterface, ok := c.Get("baseUserID")
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }
    merchantID, ok := merchantIDInterface.(uint)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "商户ID类型错误"})
        return
    }
    currentDate := time.Now().Format("2006-01-02")
    cacheKey := fmt.Sprintf("business-data-order:merchant_id:%d:date:%s", merchantID, currentDate)
    // 尝试从 Redis 获取缓存的数据
    var cachedData struct {
        WaitingOrders   int `json:"waitingOrders"`
        DeliveredOrders int `json:"deliveredOrders"`
        CompletedOrders int `json:"completedOrders"`
        CancelledOrders int `json:"cancelledOrders"`
        AllOrders       int `json:"allOrders"`
    }
    found, err := utils.GetJSON(context.Background(), cacheKey, &cachedData)
    if err != nil {
        log.Printf("Redis读取错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    if found {
        c.JSON(http.StatusOK, gin.H{
            "code": 1,
            "data": cachedData,
        })
        return
    }
    // 查询订单状态
    var orderStatusCounts []struct {
        Status int `json:"status"`
        Count  int `json:"count"`
    }
    if err := global.Db.Model(&models.Order{}).
        Where("merchant_id = ? AND DATE(pickup_point) = ?", merchantID, currentDate).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&orderStatusCounts).Error; err != nil {
        log.Printf("查询订单状态失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
    // 初始化计数变量
    var waitingCount, deliveringCount, completedCount, cancelledCount int
    // 遍历查询结果并设置计数变量
    for _, orderStatusCount := range orderStatusCounts {
        switch orderStatusCount.Status {
        case 2:
            waitingCount = orderStatusCount.Count
        case 3:
            deliveringCount = orderStatusCount.Count
        case 5:
            completedCount = orderStatusCount.Count
        case 6:
            cancelledCount = orderStatusCount.Count
        }
    }
    // 计算所有订单总数
    allOrders := waitingCount + deliveringCount + completedCount + cancelledCount
    // 准备返回数据
    data := gin.H{
        "waitingOrders":   waitingCount,
        "deliveredOrders": deliveringCount,
        "completedOrders": completedCount,
        "cancelledOrders": cancelledCount,
        "allOrders":       allOrders,
    }
    // 序列化数据并存入 Redis
    err = utils.SetJSON(context.Background(), cacheKey, data, 5*time.Minute)
    if err != nil {
        log.Printf("Redis写入错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": data,
    })
}

// GetOverviewDishes 获取菜品一览
func GetOverviewDishes(c *gin.Context) {
   baseUserID, ok := c.Get("baseUserID")
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }
    merchantID, ok := baseUserID.(uint)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "商户ID类型错误"})
        return
    }
    cacheKey := fmt.Sprintf("overview-dishes:merchant_id:%d", merchantID)
    // 尝试从 Redis 获取缓存的数据
    var cachedData struct {
        Sold         int64 `json:"sold"`
        Discontinued int64 `json:"discontinued"`
    }
    found, err := utils.GetJSON(context.Background(), cacheKey, &cachedData)
    if err != nil {
        log.Printf("Redis读取错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    if found {
        c.JSON(http.StatusOK, gin.H{
            "code": 1,
            "data": cachedData,
        })
        return
    }
    // 统计启售和停售的数量
    var soldCount, discontinuedCount int64
    // 查询 dish 表，统计启售的数量
    if err := global.Db.Model(&models.Dish{}).
        Where("merchant_id = ? AND status = ?", merchantID, 1).
        Count(&soldCount).Error; err != nil {
        fmt.Println(err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
    // 查询 dish 表，统计停售的数量
    if err := global.Db.Model(&models.Dish{}).
        Where("merchant_id = ? AND status = ?", merchantID, 0).
        Count(&discontinuedCount).Error; err != nil {
        fmt.Println(err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
    // 准备返回数据
    data := gin.H{
        "sold":         soldCount,
        "discontinued": discontinuedCount,
    }
    // 序列化数据并存入 Redis
    err = utils.SetJSON(context.Background(), cacheKey, data, 5*time.Minute)
    if err != nil {
        log.Printf("Redis写入错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": data,
    })
}

// GetOverviewDishes 获取套餐一览
func GetOverviewMeals(c *gin.Context) {
  value, ok := c.Get("baseUserID")
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }
    merchantID, ok := value.(uint)
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "商户ID类型错误"})
        return
    }
    currentDate := time.Now().Format("2006-01-02")
    cacheKey := fmt.Sprintf("business-data-meal:merchant_id:%d:date:%s", merchantID, currentDate)
    // 尝试从 Redis 获取缓存的数据
    var cachedData struct {
        Sold         int `json:"sold"`
        Discontinued int `json:"discontinued"`
    }
    found, err := utils.GetJSON(context.Background(), cacheKey, &cachedData)
    if err != nil {
        log.Printf("Redis读取错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    if found {
        c.JSON(http.StatusOK, gin.H{
            "code": 1,
            "data": cachedData,
        })
        return
    }
    // 查询 meal 表，统计启售和停售的数量
    var mealStatusCounts []struct {
        Status int `json:"status"`
        Count  int `json:"count"`
    }
    if err := global.Db.Model(&models.Meal{}).
        Where("merchant_id = ?", merchantID).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&mealStatusCounts).Error; err != nil {
        log.Printf("查询订单状态失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
    // 初始化计数变量
    var soldCount, discontinuedCount int
    // 遍历查询结果并设置计数变量
    for _, mealStatusCount := range mealStatusCounts {
        switch mealStatusCount.Status {
        case 1:
            soldCount = mealStatusCount.Count
        case 0:
            discontinuedCount = mealStatusCount.Count
        }
    }
    // 准备返回数据
    data := gin.H{
        "sold":         soldCount,
        "discontinued": discontinuedCount,
    }
    // 序列化数据并存入 Redis
    err = utils.SetJSON(context.Background(), cacheKey, data, 5*time.Minute)
    if err != nil {
        log.Printf("Redis写入错误: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "服务器内部错误"})
        return
    }
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": data,
    })
}