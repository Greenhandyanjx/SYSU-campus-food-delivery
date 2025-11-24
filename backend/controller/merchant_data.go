package controller

//商家端的数据统计相关接口
import (
	"backend/global"
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	// 查询营业额
	var revenues []models.Revenue
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

	// 提取日期和营业额列表
	dateList := make([]string, len(revenues))
	turnoverList := make([]float64, len(revenues))
	for i, revenue := range revenues {
		dateList[i] = revenue.Date.Format("2006-01-02")
		turnoverList[i] = revenue.Revenue
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code": "1",
		"msg":  "获取概览成功",
		"data": gin.H{
			"dateList":     strings.Join(dateList, ","),
			"turnoverList": turnoverList,
		},
	})
}

func GetOrderStatistics(ctx *gin.Context) {
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

	// 定义用于接收统计结果的结构体
	type OrderCountStat struct {
		Date            time.Time `gorm:"column:date"`
		TotalOrderCount int       `gorm:"column:total_order_count"`
		ValidOrderCount int       `gorm:"column:valid_order_count"`
	}
	// 执行带有聚合函数的查询
	var orderCountStats []OrderCountStat
	if err := global.Db.Table("orders").
		Select("DATE(dropof_point) as date, COUNT(id) as total_order_count, SUM(CASE WHEN status = 5 THEN 1 ELSE 0 END) as valid_order_count").
		Where("merchant_id = ? AND DATE(dropof_point) BETWEEN ? AND ?", merchantID, beginTime, endTime).
		Group("DATE(dropof_point)").
		Find(&orderCountStats).Error; err != nil {
		log.Printf("查询订单统计失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": "查询失败",
		})
		return
	}
	/// 初始化总的订单数量和有效的订单数量
	var totalOrderCount, validOrderCount int

	// 将统计结果转换为需要的格式
	dateList := make([]string, len(orderCountStats))
	orderCountList := make([]int, len(orderCountStats))
	validOrderCountList := make([]int, len(orderCountStats))

	for i, stat := range orderCountStats {
		dateList[i] = stat.Date.Format("2006-01-02")
		orderCountList[i] = stat.TotalOrderCount
		validOrderCountList[i] = stat.ValidOrderCount
		totalOrderCount += stat.TotalOrderCount
		validOrderCount += stat.ValidOrderCount
	}

	// 计算订单完成率
	orderCompletionRate := 0.0
	if totalOrderCount > 0 {
		orderCompletionRate = float64(validOrderCount) / float64(totalOrderCount)
	}

	// 将数值切片转换为逗号分隔的字符串
	orderCountStr := make([]string, len(orderCountList))
	validOrderCountStr := make([]string, len(validOrderCountList))
	for i, count := range orderCountList {
		orderCountStr[i] = strconv.Itoa(count)
	}
	for i, count := range validOrderCountList {
		validOrderCountStr[i] = strconv.Itoa(count)
	}

	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code": 1, // 注意：这里应该是数字1，不是字符串"1"
		"data": gin.H{
			"dateList":            strings.Join(dateList, ","),
			"orderCountList":      strings.Join(orderCountStr, ","),
			"validOrderCountList": strings.Join(validOrderCountStr, ","),
			"totalOrderCount":     totalOrderCount,
			"validOrderCount":     validOrderCount,
			"orderCompletionRate": orderCompletionRate,
		},
	})
}

func GetUserData(ctx *gin.Context) {
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

	// 查询总用户数和新用户数
	type UserCount struct {
		Date       string `json:"date"`
		TotalUsers int    `json:"total_users"`
	}

	var userCounts []UserCount
	if err := global.Db.Table("orders").
		Where("merchant_id = ? AND DATE(dropof_point) BETWEEN ? AND ?", merchantID, beginTime, endTime).
		Select("DATE(dropof_point) as date", "COUNT(DISTINCT consignee) as total_users").
		Group("DATE(dropof_point)").
		Find(&userCounts).Error; err != nil {
		log.Printf("查询用户数量失败: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"code":    0,
			"message": "查询失败",
		})
		return
	}

	// 检查查询结果是否为空
	if len(userCounts) == 0 {
		log.Printf("查询结果为空")
	}

	// 提取日期、总用户
	dateList := make([]string, len(userCounts))
	totalUserList := make([]int, len(userCounts))

	for i, userCount := range userCounts {
		dateList[i] = userCount.Date
		totalUserList[i] = userCount.TotalUsers
	}

	// 将切片转换为字符串
	dateListStr := strings.Join(dateList, ",")
	totalUserListStr := strings.Join(intSliceToStringSlice(totalUserList), ",")
	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code": "1",
		"data": gin.H{
			"dateList":      dateListStr,
			"totalUserList": totalUserListStr,
		},
	})
}

// 辅助函数：将 int 切片转换为 string 切片
func intSliceToStringSlice(ints []int) []string {
	strSlice := make([]string, len(ints))
	for i, num := range ints {
		strSlice[i] = strconv.Itoa(num)
	}
	return strSlice
}

// GetTopSales 返回 top N 销量项
// 请求参数：type=dish|meal, date=YYYY-MM-DD, limit=10
func GetTopSales(c *gin.Context) {
	// typ := c.Query("type")
	// if typ != "dish" && typ != "meal" {
	//     c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "type must be 'dish' or 'meal'"})
	//     return
	// }
	typ := "dish"
	// 绑定请求参数
	var params GetStatisticsParams
	if err := c.ShouldBindQuery(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "请求参数错误",
		})
		return
	}

	// 解析 begin 和 end 参数为时间格式
	beginTime, err := time.Parse("2006-01-02", params.Begin)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "begin 参数格式错误",
		})
		return
	}

	endTime, err := time.Parse("2006-01-02", params.End)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "end 参数格式错误",
		})
		return
	}

	// 确保 endTime 不早于 beginTime
	if endTime.Before(beginTime) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "end 参数不能早于 begin 参数",
		})
		return
	}

	// limitStr := c.DefaultQuery("limit", "10")
	limitStr := "10"
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	// merchant id 优先从中间件获取
	var merchantID uint
	if v, ok := c.Get("baseUserID"); ok {
		switch id := v.(type) {
		case uint:
			merchantID = id
		case int:
			merchantID = uint(id)
		}
	}
	if merchantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "merchant id required"})
		return
	}

	var stats []models.SalesStat
	if err := global.Db.Table("sales-stats").
		Where("merchant_id = ? AND item_type = ? AND date BETWEEN  ? AND ?", merchantID, typ, beginTime, endTime).
		Order("quantity desc").
		Limit(limit).
		Find(&stats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "query failed"})
		fmt.Println("error:", err)
		return
	}

	// 返回 item_id, item_name,quantity, revenue 列表
	// 提取菜品名和销量列表
	nameList := make([]string, len(stats))
	numberList := make([]int, len(stats))
	for i, s := range stats {
		nameList[i] = s.Itemname
		numberList[i] = s.Quantity
	}
	// 将切片转换为字符串
	nameListStr := strings.Join(nameList, ",")
	numberListStr := strings.Join(intSliceToStringSlice(numberList), ",")
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"nameList":   nameListStr,
			"numberList": numberListStr,
		},
	})
}
