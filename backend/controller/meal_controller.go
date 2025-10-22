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

func Meal_add(ctx *gin.Context) {
    var meal models.Meal
    var dishIDs []int

    // 绑定 JSON 数据
    if err := ctx.ShouldBindJSON(&meal); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "binding error",
        })
        return
    }

    // 获取 dish_ids
    if err := ctx.ShouldBindJSON(&dishIDs); err != nil {
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "binding dish_ids error",
        })
        return
    }

    // 处理图片上传
    file, err := ctx.FormFile("image")
    if err == nil {
        // 生成唯一文件名
        ext := filepath.Ext(file.Filename)
        timestamp := time.Now().Unix()
        filename := "meal_" + strconv.FormatInt(timestamp, 10) + ext
        relativePath := filepath.Join("uploads", "meals", filename)
        fullPath := filepath.Join(".", relativePath)
        
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
        meal.ImagePath = relativePath
    } else {
        meal.ImagePath = "" // 如果没有上传图片，ImagePath 设置为空
    }

    // 创建 meal
	if err := global.Db.Table("meals").AutoMigrate(&models.Meal{}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table create error",
		})
		return
	}
    if err := global.Db.Create(&meal).Error; err != nil {
        log.Printf("数据库创建错误: %v", err) // 记录详细错误日志
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "服务器内部错误，请稍后再试",
        })
        return
    }

    // 创建关联关系
    for _, dishID := range dishIDs {
        mealDish := models.MealDish{
            MealID:  meal.ID,
            DishID:  dishID,
        }
        if err := global.Db.Create(&mealDish).Error; err != nil {
            log.Printf("创建关联关系错误: %v", err) // 记录详细错误日志
            ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
                "code": "500",
                "msg":  "服务器内部错误，请稍后再试",
            })
            return
        }
    }

    ctx.JSON(http.StatusOK, gin.H{
        "code": "200",
        "msg":  "meal 创建成功",
    })
}
