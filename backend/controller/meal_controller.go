package controller

import (
	"backend/global"
	"backend/models"
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Meal_add(ctx *gin.Context) {
    var meal models.Meal
    // 中间结构体用于绑定
    type mealRequest struct {
        DishIDs []int `json:"dishids"`
    }
    var request mealRequest
    body, _ := io.ReadAll(ctx.Request.Body)
    fmt.Println("Request Body:", string(body)) // 打印请求体内容
    ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置请求体
    if err := ctx.ShouldBindJSON(&meal); err != nil {
        // 打印绑定错误的详细信息
		log.Printf("绑定错误: %v", err.Error())
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "binding error",
        })
        return
    }
    fmt.Println("Request Body:", string(body)) // 打印请求体内容
   	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
    // 获取 dish_ids
    if err := ctx.ShouldBindJSON(&request); err != nil {
        // 打印绑定错误的详细信息
		log.Printf("绑定错误: %v", err.Error())
        ctx.JSON(http.StatusBadRequest, gin.H{
            "code": "400",
            "msg":  "binding dish_ids error",
        })
        return
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
        ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "服务器内部错误，请稍后再试",
        })
        return
    }

    // 创建关联关系
    if err := global.Db.Table("sysu_campus_food.meal_dishes").AutoMigrate(&models.MealDish{}); err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg": "table md create error",
		})
		return
	}
    dishIDs := request.DishIDs
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

func QueryDishById(ctx *gin.Context) {
    id := ctx.Query("id")
    var dish models.Dish
    if err := global.Db.Model(&models.Dish{}).Where("ID = ?", id).First(&dish).Error; err != nil {
        ctx.JSON(http.StatusInternalServerError, gin.H{
            "code": "500",
            "msg":  "查询菜品失败",
        })
        return
    }
    ctx.JSON(http.StatusOK, gin.H{
        "code": "1",
        "msg":  "success",
        "data": dish,
    })
}
