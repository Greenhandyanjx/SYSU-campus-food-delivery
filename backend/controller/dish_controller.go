package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func Dish_add(ctx*gin.Context){
    var dish models.Dish
    if err := ctx.ShouldBind(&dish); err!= nil {
		ctx.JSON(http.StatusBadRequest,gin.H{
			"code":"400",
			"msg":"binding error",
		})
	}

	var imagePath string
	
	// 处理图片上传
	file, err := ctx.FormFile("image")
	if err == nil {
		// 生成唯一文件名
		ext := filepath.Ext(file.Filename)
		timestamp := time.Now().Unix()
		filename := "meal_" + strconv.FormatInt(timestamp, 10) + ext
		baseUploadPath := global.Meal_image_path // 绝对路径
        relativePath := filepath.Join(baseUploadPath, filename)
        fullPath := relativePath // 直接使用相对路径作为完整路径
		// 确保上传目录存在
		if err := utils.EnsureDir(fullPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法创建上传目录"})
			return
		}
		// 保存文件
		if err := ctx.SaveUploadedFile(file, fullPath); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "无法保存图片"})
			return
		}
		imagePath = relativePath
	}
	dish.ImagePath = imagePath
	if err := global.Db.Table("dishes").AutoMigrate(&dish); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return
	}
	if err := global.Db.Create(&dish).Error; err != nil {
		log.Printf("数据库创建错误: %v", err) // 记录详细错误日志
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}
}

