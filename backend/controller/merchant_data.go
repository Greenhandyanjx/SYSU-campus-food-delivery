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

type GetStatisticsParams struct {
    Begin string `form:"begin" json:"begin"`
    End   string `form:"end" json:"end"`
}

func GetDataOverView(ctx *gin.Context) {
    // 获取上下文中的 baseUserID
    baseUserID, exists := ctx.Get("baseUserID")
    if !exists {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "code": "401",
            "msg":  "未找到商户ID",
        })
        return
    }

    // 确保 baseUserID 是 uint 类型
    merchantID, ok := baseUserID.(uint)
    if !ok {
        ctx.JSON(http.StatusUnauthorized, gin.H{
            "code": "401",
            "msg":  "商户ID类型错误",
        })
        return
    }

    // 绑定请求参数
    var params GetStatisticsParams
    if err := ctx.ShouldBindQuery(&params); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "请求参数错误",
        })
        return
    }

    // 解析 begin 和 end 参数为时间格式
    beginTime, err := time.Parse("2006-01-02", params.Begin)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "begin 参数格式错误",
        })
        return
    }

    endTime, err := time.Parse("2006-01-02", params.End)
    if err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "end 参数格式错误",
        })
        return
    }

    // 确保 endTime 不早于 beginTime
    if endTime.Before(beginTime) {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "end 参数不能早于 begin 参数",
        })
        return
    }

    // 初始化统计变量
    var validOrdersCount, userCount int64
    type Revenue struct {
    Date    string  `json:"date"`
    Revenue float64 `json:"revenue"`
   }
   // 查询营业额
var revenues []Revenue
if err := global.Db.Table("revenues").
    Where("merchant_id = ? AND date BETWEEN ? AND ?", merchantID, beginTime, endTime).
    Select("date", "revenue").
    Find(&revenues).Error; err != nil {
    log.Printf("查询营业额失败: %v", err)
    ctx.JSON(http.StatusInternalServerError, gin.H{
        "code":    0,
        "message": "查询失败",
    })
    return
}
// 检查查询结果是否为空
if len(revenues) == 0 {
    log.Printf("查询结果为空")
}
// 打印结果
fmt.Println(revenues)
    // 查询有效订单数量 (假设有效订单是指 status >= 2 且 status <= 5 的订单)
    if err := global.Db.Model(&models.Order{}).
        Where("merchant_id = ? AND Date(dropof_point) BETWEEN ? AND ? AND status >= ? AND status <= ?", merchantID, beginTime, endTime, 2, 5).
        Count(&validOrdersCount).Error; err != nil {
        log.Printf("查询有效订单数量失败: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "code":    0,
            "message": "查询失败",
        })
        return
    }

    // 查询用户数量 (假设用户是指下单的用户)
    if err := global.Db.Model(&models.Order{}).
        Where("merchant_id = ? AND Date(dropof_point) BETWEEN ? AND ?", merchantID, beginTime, endTime).
        Select("Distinct consignee").
        Count(&userCount).Error; err != nil {
        log.Printf("查询用户数量失败: %v", err)
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "code":    0,
            "message": "查询失败",
        })
        return
    }
    // 返回结果
    ctx.JSON(http.StatusOK, gin.H{
        "code": "1",
        "msg":  "获取概览成功",
        "data": gin.H{
            "turnover":        revenues,
            "validOrdersCount": validOrdersCount,
            "userCount":        userCount,
        },
    })
}
