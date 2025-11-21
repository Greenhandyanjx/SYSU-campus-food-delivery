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
		noAuth.POST("/register", controller.Register)
		// 兼容以前的单独注册路由，指向统一的 Register
		noAuth.POST("/register/user", controller.Register)
		noAuth.POST("/register/rider", controller.Register)
		noAuth.POST("/register/merchant", controller.Register)

		noAuth.GET("/merchant/category/list", controller.Get_category)
		noAuth.GET("/merchant/dish/categories", controller.Get_category)

		noAuth.GET("/merchant/common/download", controller.CommonDownload)
		// 允许未登录用户上传图片（用于注册页面等）
		noAuth.POST("/common/upload", controller.UploadImage)

		noAuth.GET("/merchant/orders/status", controller.GetOrderListByStatus)
		noAuth.GET("/merchant/orders/page", controller.GetOrderPage)
		noAuth.GET("/merchant/order/detail", controller.GetOrderDetail)
		// lookup endpoints for frontend
		noAuth.GET("/merchant/detail", controller.GetMerchantDetail)
		noAuth.GET("/baseuser/detail", controller.GetBaseUserDetail)
		// debug: list active websocket connections (base_user ids)
		noAuth.GET("/debug/ws/connections", controller.DebugConnections)
		// WebSocket 握手允许通过查询参数 token 或 uid 进行鉴权，放到无鉴权组以便控制器自行处理
		noAuth.GET("/chat/ws", controller.ChatWS)
		noAuth.POST("/merchant/order/add", controller.Orderadd)

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
		auth.GET("/merchant/dishes/page", controller.Get_dishes)

		auth.POST("/merchant/meal/add", controller.Meal_add)
		auth.POST("/merchant/meal/status", controller.Edit_Meal_Status)
		auth.POST("/merchant/meal/delete", controller.Meal_Delete)
		auth.POST("merchant/meal/edit", controller.Meal_Edit)
		auth.GET("/merchant/meal/query", controller.Get_Meal_ById)
		auth.GET("/merchant/meal/page", controller.GetMealsPage)

		auth.POST("/merchant/order/accept", controller.OrderAccept)
		auth.POST("/merchant/order/reject", controller.OrderReject)
		auth.POST("/merchant/order/delivery", controller.OrderDelivery)
		auth.POST("/merchant/order/complete", controller.OrderComplete)
		// 其他需要中间件保护的路由可以添加在这里
		// 聊天 WebSocket 与历史接口
		auth.GET("/chat/history", controller.ChatHistory)
		// 用户侧会话列表
		auth.GET("/user/chats", controller.GetUserChats)
		// 用户侧已读标记
		auth.POST("/user/chats/mark_read", controller.MarkUserChatRead)
		// 商家会话列表与已读标记
		auth.GET("/merchant/chats", controller.GetMerchantChats)
		auth.POST("/merchant/chats/mark_read", controller.MarkChatRead)
		//需要中间件读取的信息
		auth.GET("/merchant/businessData", controller.GetBusinessData)
		auth.GET("/merchant/orderData", controller.GetOrderData)
		auth.GET("/merchant/overviewDishes", controller.GetOverviewDishes)
		auth.GET("/merchant/setMealStatistics", controller.GetOverviewMeals)
		//数据统计页面
		auth.GET("/merchant/statistics/turnover", controller.GetDataOverView)
		auth.GET("/merchant/statistics/user", controller.GetUserData)
		auth.GET("/merchant/statistics/order", controller.GetOrderStatistics)
		// ====== Rider APIs ======
		auth.GET("/rider/info", controller.GetRiderInfo)
		auth.POST("/rider/status", controller.UpdateRiderStatus)
		auth.GET("/rider/orders/new", controller.GetNewOrders)
		auth.POST("/rider/orders/:orderId/accept", controller.AcceptOrder)
		auth.POST("/rider/orders/:orderId/pickup", controller.PickupOrder)
		auth.GET("/rider/orders/delivering", controller.GetDeliveringOrders)
		auth.POST("/rider/orders/:orderId/complete", controller.CompleteOrder)
		auth.GET("/rider/orders/history", controller.GetOrderHistory)

		auth.GET("merchant/statistics/top", controller.GetTopSales)
	}
	return fe
}
