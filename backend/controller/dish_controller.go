package controller

import (
	"backend/global"
	"backend/models"
	"log"
	"net/http"

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

