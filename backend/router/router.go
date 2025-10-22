package router

import (
	"backend/controller"
	"backend/midware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
		fe := gin.Default()
	fe.Use(cors.Default())
	// 创建一个不需要中间件的组
	noAuth := fe.Group("/api")
	{
		noAuth.POST("/login", controller.Login)
		noAuth.POST("user/register", controller.Register_User)
		noAuth.POST("rider/register", controller.Register_Rider)
		noAuth.POST("merchant/register", controller.Register_Merchant)
	}
	// 创建一个需要中间件的组
	auth:=fe.Group("/api")
	auth.Use(midware.AuthMiddleware()) 
	{
		auth.POST("/change_password", controller.ChangePassword)
		auth.POST("/merchant/dish/add",controller.Dish_add)
		auth.POST("/merchant/meal/add",controller.Meal_add)
		// 其他需要中间件保护的路由可以添加在这里
	}
	return fe
}