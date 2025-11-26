package utils

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)
type GetStatisticsParams struct {
    Begin string `form:"begin" binding:"required"`
    End   string `form:"end" binding:"required"`
}

func MerchantStatsParams(ctx *gin.Context) (uint, time.Time, time.Time, error) {
	// 从 context 中获取商户ID
	baseUserID, exists := ctx.Get("baseUserID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": "401",
			"msg":  "未找到商户ID",
		})
		return 0, time.Time{}, time.Time{}, fmt.Errorf("未找到商户ID")
	}

	// 确保 baseUserID 是 uint 类型
	merchantID, ok := baseUserID.(uint)
	if !ok {
		ctx.JSON(http.StatusUnauthorized, gin.H{
			"code": "401",
			"msg":  "商户ID类型错误",
		})
		return 0, time.Time{}, time.Time{}, fmt.Errorf("商户ID类型错误")
	}

	// 绑定请求参数
	var params GetStatisticsParams
	if err := ctx.ShouldBindQuery(&params); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "请求参数错误",
		})
		return 0, time.Time{}, time.Time{}, fmt.Errorf("请求参数错误: %v", err)
	}

	// 解析 begin 和 end 参数为时间格式
	beginTime, err := time.Parse("2006-01-02", params.Begin)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "begin 参数格式错误",
		})
		return 0, time.Time{}, time.Time{}, fmt.Errorf("begin 参数格式错误: %v", err)
	}

	endTime, err := time.Parse("2006-01-02", params.End)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "end 参数格式错误",
		})
		return 0, time.Time{}, time.Time{}, fmt.Errorf("end 参数格式错误: %v", err)
	}

	// 确保 endTime 不早于 beginTime
	if endTime.Before(beginTime) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "end 参数不能早于 begin 参数",
		})
		return 0, time.Time{}, time.Time{}, fmt.Errorf("end 参数不能早于 begin 参数")
	}

	return merchantID, beginTime, endTime, nil
}
