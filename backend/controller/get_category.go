package controller

import (
	"backend/global"
	"backend/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Get_category(ctx *gin.Context){
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
    // 如果有类型参数，则添加类型过滤条件
    if categoryType != "" {
        query = query.Where("type = ?", categoryType)
    }
    // 查询分页数据
    if err := query.Offset(offset).Limit(size).Find(&categories).Error; err != nil {
        log.Printf("数据库查询错误: %v", err) // 记录详细错误日志
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "服务器内部错误，请稍后再试",
        })
        return
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
        "code": "200",
        "msg":  "查询成功",
        "data": gin.H{
            "items": categories,
            "total": total,
        },
    })
}