package controller

import (
	"backend/global"
	"backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Get_category(ctx *gin.Context) {
	// 获取请求参数
	categoryType := ctx.Query("type")
	pageStr := ctx.DefaultQuery("page", "1")
	sizeStr := ctx.DefaultQuery("size", "10")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "无效的页码参数",
		})
		return
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "无效的页大小参数",
		})
		return
	}
	// 计算分页偏移量
	offset := (page - 1) * size
	// 定义一个切片来存储查询结果
	var categories []models.Category
	// 构建查询条件
	query := global.Db.Model(&models.Category{}).Select("id", "name", "type", "sort")
	// 如果有类型参数，则按约定返回对应分类集合：
	// - type=="1" (菜品)：返回 ID 1..15（所有固定分类）
	// - type=="2" (套餐)：返回用于套餐的分类（招牌套餐:id=1, 午餐推荐:id=12, 低价满减:id=13）
	if categoryType != "" {
		if categoryType == "1" {
			query = query.Where("id >= ? AND id <= ?", 1, 15)
		} else if categoryType == "2" {
			query = query.Where("id IN ?", []int{1, 12, 13})
		} else {
			query = query.Where("type = ?", categoryType)
		}
	}
	// 查询分页数据
	// 对于固定的菜品分类 (type=="1") 与套餐推荐 (type=="2"), 返回完整集合而不是默认分页
	if categoryType == "1" || categoryType == "2" {
		if err := query.Find(&categories).Error; err != nil {
			log.Printf("数据库查询错误: %v", err) // 记录详细错误日志
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": "500",
				"msg":  "服务器内部错误，请稍后再试",
			})
			return
		}
	} else {
		if err := query.Offset(offset).Limit(size).Find(&categories).Error; err != nil {
			log.Printf("数据库查询错误: %v", err) // 记录详细错误日志
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": "500",
				"msg":  "服务器内部错误，请稍后再试",
			})
			return
		}
	}
	// 查询总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("数据库计数错误: %v", err) // 记录详细错误日志
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}
	// 返回查询结果
	ctx.JSON(http.StatusOK, gin.H{
		"code":  1,
		"msg":   "success",
		"data":  categories,
		"total": total,
	})
}

func Get_category_by_merchantid(ctx *gin.Context) {
	merchantId := ctx.Query("merchantId")
	var resultCategories []models.Category
	// 构建查询条件
	var query *gorm.DB
	if merchantId == "" {
		// 如果没有提供merchantId，则返回所有分类
		query = global.Db.Find(&resultCategories)
	} else {
		// 如果提供了merchantId，则根据merchantId筛选分类
		query = global.Db.Where("id = ?", merchantId).Find(&resultCategories)
	}
	// 检查查询是否成功
	if query.Error != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询分类列表失败", "data": nil})
		return
	}
	// 返回结果
	ctx.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": resultCategories,
	})
}
