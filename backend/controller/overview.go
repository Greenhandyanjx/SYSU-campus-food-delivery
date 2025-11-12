package controller

import (
	"backend/global"
	"backend/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetBusinessData 获取店铺的经营概览
func GetBusinessData(c *gin.Context) {
    var request models.BusinessDataRequest

    // 解析请求体
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }    
    // 查询状态为 1 的订单，并确保 merchantid 符合 baseid
    var count int64
    var totalRevenue float64
	 // 查询 revenue 表，获取对应日期的营业额
    if err := global.Db.Model(&models.Revenue{}).
        Where("merchant_id = ? AND date = ?",request.BaseID, request.Date).
        Select("revenue").
        Scan(&totalRevenue).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }

  // 查询状态为 1 的订单数量
    if err := global.Db.Model(&models.Order{}).
        Where("status = ? AND merchant_id = ? AND DATE(pickup_point) = ?", request.BaseID,request.Date).
        Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": 0, 
		    "message": "查询失败",
	    })
        return
    }

    // 计算平均价格
    var avgTicket float64
    if count > 0 {
        avgTicket = totalRevenue / float64(count)
    }

    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "revenue":   totalRevenue,
            "orders":    count,
            "avgTicket": avgTicket,
        },
    })
}

// GetOrderData 获取当日订单统计
func GetOrderData(c *gin.Context) {
    var request models.BusinessDataRequest

    // 解析请求体
    if err := c.ShouldBindJSON(&request); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }

    // 统计各类订单数量
    var pendingCount, deliveringCount, completedCount int64

    // 查询订单状态
    if err := global.Db.Model(&models.Order{}).
        Where("merchant_id = ? AND DATE(pickup_point) = ?", request.BaseID, request.Date).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&[]struct {
            Status int64 `json:"status"`
            Count  int64 `json:"count"`
        }{
            {Status: 1, Count: pendingCount}, // 假设 1 为待处理状态
            {Status: 2, Count: deliveringCount}, // 假设 2 为配送中状态
            {Status: 3, Count: completedCount}, // 假设 3 为已完成状态
        }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }

    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "pending":   pendingCount,
            "delivering": deliveringCount,
            "completed":  completedCount,
        },
    })
}

// GetOverviewDishes 获取菜品一览
func GetOverviewDishes(c *gin.Context) {
    var id uint

    // 解析请求体
    if err := c.ShouldBindJSON(&id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }

    // 统计启售和停售的数量
    var soldCount, discontinuedCount int64

    // 查询 dish 表，统计启售和停售的数量
    if err := global.Db.Model(&models.Dish{}).
        Where("merchantid = ?", id).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&[]struct {
            Status int64 `json:"status"`
            Count  int64 `json:"count"`
        }{
            {Status: 1, Count: soldCount},        // 假设 1 为启售状态
            {Status: 0, Count: discontinuedCount}, // 假设 0 为停售状态
        }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }

    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "sold":        soldCount,
            "discontinued": discontinuedCount,
        },
    })
}

// GetOverviewDishes 获取菜品一览
func GetOverviewMeals(c *gin.Context) {
    var id uint

    // 解析请求体
    if err := c.ShouldBindJSON(&id); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }

    // 统计启售和停售的数量
    var soldCount, discontinuedCount int64

    // 查询 dish 表，统计启售和停售的数量
    if err := global.Db.Model(&models.Meal{}).
        Where("merchantid = ?", id).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&[]struct {
            Status int64 `json:"status"`
            Count  int64 `json:"count"`
        }{
            {Status: 1, Count: soldCount},        // 假设 1 为启售状态
            {Status: 0, Count: discontinuedCount}, // 假设 0 为停售状态
        }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }

    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "sold":        soldCount,
            "discontinued": discontinuedCount,
        },
    })
}