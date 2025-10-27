package controller

import (
	"backend/global"
	"backend/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Get_category(ctx *gin.Context){
	 var categories []models.Category
    // 从数据库中查询所有 category 的 name 属性
    if err := global.Db.Model(&models.Category{}).Select("name").Find(&categories).Error; err != nil {
        log.Printf("数据库查询错误: %v", err) // 记录详细错误日志
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "服务器内部错误，请稍后再试",
        })
        return
    }
    // 提取 name 属性到一个新的切片
    var categoryNames []string
    for _, category := range categories {
        categoryNames = append(categoryNames, category.Name)
    }
    // 返回查询结果
    ctx.JSON(http.StatusOK, gin.H{
        "code": "200",
        "msg":  "查询成功",
        "data": categoryNames,
    })
}