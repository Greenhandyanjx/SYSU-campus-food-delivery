package controller

import (
	"backend/global"
	"backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdateRevenue 更新营业额
func UpdateRevenue(c *gin.Context) {
    var order models.Order // 假设你已经有了 Order 结构体
    if err := c.ShouldBindJSON(&order); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数错误"})
        return
    }

    // 检查订单状态
    if order.Status != 1 { // 假设 1 表示订单完成
        c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "订单未完成"})
        return
    }

    // 获取当前日期
    currentDate := time.Now().Truncate(24 * time.Hour) // 只保留日期部分

    // 查找或创建营业额记录
    var revenue models.Revenue
    if err := global.Db.Where("merchant_id = ? AND date = ?", order.MerchantID, currentDate).First(&revenue).Error; err != nil {
        // 如果没有找到，创建新的记录
        revenue = models.Revenue{
            MerchantID: order.MerchantID,
            Revenue:    order.TotalPrice, // 假设订单中有 Price 字段
            Day:       currentDate,
        }
        if err := global.Db.Create(&revenue).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新营业额失败"})
            return
        }
    } else {
        // 如果找到，更新营业额
        revenue.Revenue += order.TotalPrice // 增加营业额
        if err := global.Db.Save(&revenue).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新营业额失败"})
            return
        }
    }

    c.JSON(http.StatusOK, gin.H{"code": 1, "message": "营业额更新成功"})
}