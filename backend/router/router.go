package router

import (
	"backend/controller"
	"backend/global"
	"backend/midware"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetRouter() *gin.Engine {
	fe := gin.Default()
	fe.Static("/images", global.Meal_image_path) // 静态文件服务，用于访问上传的图片
	fe.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // 前端地址
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// 创建一个不需要中间件的组
	noAuth := fe.Group("/api")
	{
		noAuth.POST("/login", controller.Login)
		noAuth.POST("/register", controller.Register_Base_User_temp)
		noAuth.POST("/user/register", controller.Register_User)
		noAuth.POST("/rider/register", controller.Register_Rider)
		noAuth.POST("/merchant/register", controller.Register_Merchant)

		noAuth.GET("/merchant/category/list", controller.Get_category)
		noAuth.GET("/merchant/dish/categories", controller.Get_category)
		noAuth.GET("/merchant/dishes/page", controller.Get_dishes)

		noAuth.GET("/merchant/common/download", controller.CommonDownload)

		noAuth.GET("/merchant/orders/status",controller.GetOrderListByStatus)
		noAuth.GET("/merchant/orders/page",controller.GetOrderPage)
		noAuth.GET("/merchant/order/detail",controller.GetOrderDetail)
		noAuth.POST("/merchant/order/add",controller.Orderadd)
        noAuth.GET("/merchant/businessData",controller.GetBusinessData)
		noAuth.GET("/merchant/orderData", controller.GetOrderData)
		noAuth.GET("/merchant/overviewDishes",controller.GetOverviewDishes)
		noAuth.GET("/merchant/setMealStatistics",controller.GetOverviewMeals)
	}
	// 创建一个需要中间件的组
	auth := fe.Group("/api")
	auth.Use(midware.AuthMiddleware())
	{
		auth.POST("/change_password", controller.ChangePassword)

		auth.POST("/merchant/dish/add", controller.Dish_add)
		auth.POST("/merchant/dish/edit", controller.Edit_dish)
		auth.POST("/merchant/dish/delete", controller.Delete_dish)
		auth.GET("/merchant/dish/list", controller.QueryDishList)
		auth.GET("/merchant/dish/query", controller.Get_Dish_ById)
		auth.POST("/merchant/dish/status", controller.Edit_DishStatus_By_Status)

		auth.POST("/common/upload", controller.UploadImage)

		auth.POST("/merchant/meal/add", controller.Meal_add)
		auth.POST("/merchant/meal/status", controller.Edit_Meal_Status)
		auth.POST("/merchant/meal/delete", controller.Meal_Delete)
		auth.POST("merchant/meal/edit", controller.Meal_Edit)
		auth.GET("/merchant/meal/query", controller.Get_Meal_ById)
		auth.GET("/merchant/meal/page", controller.GetMealsPage)
		auth.POST("/merchant/order/accept",controller.OrderAccept)
		auth.POST("/merchant/order/reject",controller.OrderReject)
		// 其他需要中间件保护的路由可以添加在这里
	}
	return fe
}
