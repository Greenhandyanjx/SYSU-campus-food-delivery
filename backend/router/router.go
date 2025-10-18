package router

import (
	"backend/controller"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
	fe := gin.Default()
	fe.Use(cors.Default())
	auth := fe.Group("/api")
	{
		auth.POST("/login", controller.Login)
		auth.POST("user/register", controller.Register_User)
		auth.POST("rider/register", controller.Register_Rider)
		auth.POST("merchant/register", controller.Register_Merchant)
	}
	return fe
}