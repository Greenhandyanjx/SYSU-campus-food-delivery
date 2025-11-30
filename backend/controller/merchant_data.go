package controller

//商家端的数据统计相关接口
import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)


func GetDataOverView(ctx *gin.Context) {
	// 获取上下文中的 baseUserID
	merchantID, beginTime, endTime, err := utils.MerchantStatsParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}
	// 查询营业额
	var revenues []models.Revenue
	if err := global.Db.Table("revenues").
		Where("merchant_id = ? AND day BETWEEN ? AND ?", merchantID, beginTime, endTime).
		Select("day", "revenue").
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
		dateList[i] = revenue.Day.Format("2006-01-02")
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
	//参数处理
	merchantID, beginTime, endTime, err := utils.MerchantStatsParams(ctx)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "参数错误: " + err.Error(),
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
    merchantID, beginTime, endTime, err := utils.MerchantStatsParams(ctx)
	if err!=nil{
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
    var results []models.Revenue
	if err:= global.Db.Table("revenues").
	Where("merchant_id=? AND day BETWEEN ? AND ?",merchantID,beginTime,endTime).
	Find(&results).Error;err!=nil{
        ctx.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
        return
    }
    fmt.Println(results)
     // 初始化切片
    dateList := []string{}
    userNumbList := []string{}
    // 遍历查询结果
    for _, result := range results {
        dateList = append(dateList, result.Day.Format("2006-01-02"))
        userNumbList = append(userNumbList, strconv.Itoa(result.Usernumber))
    }
	
    // 返回结果
    ctx.JSON(http.StatusOK, gin.H{
        "data": gin.H{
            "dateList":    strings.Join(dateList, ","),
            "userNumbList": strings.Join(userNumbList, ","),
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
    merchantID, beginTime, endTime, err := utils.MerchantStatsParams(c)
    if err!=nil{
		c.JSON(http.StatusBadRequest,gin.H{
			"code":"0",
			"msg":"参数错误:"+err.Error(),
		})
	}
	// limitStr := c.DefaultQuery("limit", "10")
	limitStr := "10"
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
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
