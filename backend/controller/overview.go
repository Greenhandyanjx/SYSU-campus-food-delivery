package controller

import (
	"backend/global"
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetBusinessData 获取店铺的经营概览
func GetBusinessData(c *gin.Context) {
    baseid,ok:= c.Get("baseUserID");
    // 解析请求体
    if !ok {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }    
    var totalRevenue float64
     currentDate := time.Now().Format("2006-01-02")
	 // 查询 revenue 表，获取对应日期的营业额
    if err := global.Db.Model(&models.Revenue{}).
        Where("merchant_id = ? AND day = ?", baseid, currentDate).
        Select("revenue").
        Scan(&totalRevenue).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
  // 初始化计数变量
    var countBetween2And6, countForStatus5 int64
   // 查询状态为大于2小于6的订单数量
    if err := global.Db.Model(&models.Order{}).
        Where("status > ? AND status < ? AND merchant_id = ? AND DATE(pickup_point) = ?", 2, 6, baseid, currentDate).
        Count(&countBetween2And6).Error; err != nil {
        log.Printf("查询状态为大于2小于6的订单数量失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    0,
            "message": "查询失败",
        })
        return
    }
    // 查询状态为5的订单数量
    if err := global.Db.Model(&models.Order{}).
        Where("status = ? AND merchant_id = ? AND DATE(pickup_point) = ?", 5, baseid, currentDate).
        Count(&countForStatus5).Error; err != nil {
        log.Printf("查询状态为5的订单数量失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    0,
            "message": "查询失败",
        })
        return
    }

    //计算平均价格
    var avgTicket float64
    if countBetween2And6 > 0 {
        avgTicket = totalRevenue / float64(countBetween2And6)
    }
     fmt.Println(countForStatus5,countBetween2And6)
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
            "turnover": totalRevenue,
            "validOrderCount":countBetween2And6,
            "orderCompletionRate":float64(countForStatus5)/float64(countBetween2And6),
            "unitPrice": avgTicket,
            "newUsers":0,
        },
    })
}

// GetOrderData 获取当日订单统计
func GetOrderData(c *gin.Context) {
     merchantID,err := c.Get("baseUserID");
    // 解析请求体
    if !err {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }    
   // 查询订单状态
    var orderStatusCounts []struct {
        Status int `json:"status"`
        Count  int `json:"count"`
    }
    if err := global.Db.Model(&models.Order{}).
        Where("merchant_id = ? AND DATE(pickup_point) = ?", merchantID, time.Now().Format("2006-01-02")).
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
    // 返回结果
    c.JSON(http.StatusOK, gin.H{
        "code": 1,
        "data": gin.H{
          "waitingOrders": waitingCount,
          "deliveredOrders": deliveringCount,
          "completedOrders": completedCount,
          "cancelledOrders": cancelledCount,
          "allOrders": waitingCount + deliveringCount + completedCount + cancelledCount,
        },
    })
}

// GetOverviewDishes 获取菜品一览
func GetOverviewDishes(c *gin.Context) {
   baseUserID,err := c.Get("baseUserID");
    // 解析请求体
    if !err {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }    
     
  // 统计启售和停售的数量
    var soldCount, discontinuedCount int64
    // 查询 dish 表，统计启售和停售的数量
    if err := global.Db.Model(&models.Dish{}).
        Where("merchant_id = ? AND status = ?", baseUserID, 1).
        Count(&soldCount).Error; err != nil {
        fmt.Println(err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
    if err := global.Db.Model(&models.Dish{}).
        Where("merchant_id = ? AND status = ?", baseUserID, 0).
        Count(&discontinuedCount).Error; err != nil {
        fmt.Println(err)
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

// GetOverviewDishes 获取套餐一览
func GetOverviewMeals(c *gin.Context) {
   value,err := c.Get("baseUserID");
    // 解析请求体
    if !err {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }    

    // 查询 meal 表，统计启售和停售的数量
    // 查询订单状态
    var mealStatusCounts []struct {
        Status int `json:"status"`
        Count  int `json:"count"`
    }
    if err := global.Db.Model(&models.Meal{}).
        Where("merchant_id = ?", value).
        Select("status, COUNT(*) as count").
        Group("status").
        Scan(&mealStatusCounts).Error; err != nil {
        log.Printf("查询订单状态失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询失败"})
        return
    }
    // 初始化计数变量
    var soldCount,discontinuedCount int
    // 遍历查询结果并设置计数变量
    for _, mealStatusCount := range mealStatusCounts {
        switch mealStatusCount.Status {
        case 1:
            soldCount = mealStatusCount.Count
        case 0:
            discontinuedCount = mealStatusCount.Count
        }
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