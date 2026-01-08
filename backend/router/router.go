package router

import (
	"backend/controller"
	"backend/controller/rider"
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
		AllowOrigins:     []string{"http://localhost:5173", "http://sysu-campus-food-jiadi.site"}, // 生产环境}, // 前端地址
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
		// 管理性接口（一次性操作）：替换 categories 表（注意：需要 ?force=1）
		noAuth.POST("/admin/seed_categories", controller.ReplaceCategories)
		// 用户侧店铺查询与菜品接口
		noAuth.GET("/store/query", controller.GetStoreByQuery)
		noAuth.GET("/store/dishes", controller.GetStoreDishes)
		// 公共接口：按 meal id 查询套餐详情（包含 setmealDishes），无需鉴权，供用户端预览使用
		noAuth.GET("/store/meal/query", controller.GetPublicMealById)
		// 用户首页获取店铺列表（无需鉴权）
		noAuth.GET("/user/stores", controller.GetStores)
		// 允许未登录用户上传图片（用于注册页面等）
		noAuth.POST("/common/upload", controller.UploadImage)

		// 商家订单列表与详情需要鉴权，移动到 auth 组下面以便中间件设置 baseUserID
		// 支付平台回调（notify）
		noAuth.POST("/order/notify", controller.PaymentNotify)
		// 前端轮询订单状态（允许无鉴权以兼容扫码页面）
		noAuth.GET("/order/status", controller.GetOrderStatus)
		// lookup endpoints for frontend
		noAuth.GET("/merchant/detail", controller.GetMerchantDetail)
		// 获取商家配送配置（前端展示起送价/配送费/配送范围）
		noAuth.GET("/merchant/delivery_config", controller.GetMerchantDeliveryConfig)
		noAuth.GET("/baseuser/detail", controller.GetBaseUserDetail)
		// debug: list active websocket connections (base_user ids)
		noAuth.GET("/debug/ws/connections", controller.DebugConnections)
		// debug: send a test payload to a specific base_id ws connection
		noAuth.POST("/debug/ws/send", controller.DebugSendWS)
		// WebSocket 握手允许通过查询参数 token 或 uid 进行鉴权，放到无鉴权组以便控制器自行处理
		noAuth.GET("/chat/ws", controller.ChatWS)
		noAuth.POST("/merchant/order/add", controller.Orderadd)

	}
	// 创建一个需要中间件的组
	auth := fe.Group("/api")
	auth.Use(midware.AuthMiddleware())
	{
		// 查询收货人信息（用于商家根据 consigneeid 查找对应的 base user id）
		auth.GET("/consignee/query", controller.GetConsigneeById)
		// merchant order endpoints require authentication
		auth.GET("/merchant/orders/status", controller.GetOrderListByStatus)
		auth.GET("/merchant/orders/page", controller.GetOrderPage)
		auth.GET("/merchant/order/detail", controller.GetOrderDetail)
		auth.POST("/change_password", controller.ChangePassword)

		auth.POST("/merchant/dish/add", controller.Dish_add)
		// 商家个人信息（认证）
		auth.GET("/merchant/profile", controller.GetMerchantProfile)
		auth.POST("/merchant/profile/update", controller.UpdateMerchantProfile)
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
		auth.POST("/merchant/order/cancel", controller.OrderCancel)
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
		// ====== Rider APIs (NEW) ======
		rg := auth.Group("/rider")
		{
			rg.GET("/me", rider.GetMe)
			rg.POST("/online", rider.UpdateOnline)  // 给前端/测试用
			rg.PATCH("/online", rider.UpdateOnline) // 保留 REST 风格

			rg.GET("/orders/new", rider.GetNewOrders)
			rg.POST("/orders/:id/accept", rider.AcceptOrder)
			rg.POST("/orders/:id/pickup", rider.PickupOrder)
			rg.POST("/orders/:id/deliver", rider.DeliverOrder)

			rg.GET("/orders/ongoing", rider.GetOngoingOrders)
			rg.GET("/orders/history", rider.GetHistoryOrders)
			rg.GET("/wallet", rider.GetWallet)
			rg.GET("/income", rider.GetIncome)
			rg.POST("/withdraw", rider.ApplyWithdraw)
			rg.GET("/withdraws", rider.GetWithdraws)
			rg.POST("/location", rider.UpdateLocation)
			rg.GET("/stat", rider.GetStat)

		}

		auth.GET("merchant/statistics/top", controller.GetTopSales)

		auth.GET("/user/cart", controller.GetUserCart)
		auth.GET("/user/profile", controller.GetUserProfile)
		auth.POST("/user/profile/update", controller.UpdateUserProfile)
		auth.POST("/user/cart/add", controller.AddToCart) // 添加到购物车
		// 更新购物车（包含修改数量 / 删除，当 qty==0 时会删除该项）
		auth.POST("/user/cart/update", controller.UpdateCartItem)
		// 兼容旧的 remove 路径，指向同一个处理函数
		auth.POST("/user/cart/remove", controller.UpdateCartItem)
		auth.POST("/user/cart/deleteSelected", controller.DeleteSelected)
		auth.POST("/user/cart/selectAll", controller.SelectAll)
		auth.POST("/user/cart/selectItem", controller.SelectItem)
		auth.POST("/user/cart/selectShop", controller.SelectShop)

		// 支付相关接口
		auth.POST("/order/createPayOrder", controller.CreatePayOrder)
		// 创建 pending（用户进入结算但未完成支付时持久化的订单）
		auth.POST("/order/createPending", controller.CreatePendingOrder)

		// 用户端订单接口：列表与详情
		auth.GET("/user/order/list", controller.GetUserOrderList)
		auth.GET("/user/order/:id", controller.GetUserOrderDetail)
		// 用户评价订单
		auth.POST("/user/order/:id/review", controller.ReviewOrder)
		// 用户端订单操作：取消、支付、更新收货人/地址
		auth.POST("/user/order/cancel", controller.CancelOrder)
		auth.POST("/user/order/pay", controller.PayOrder)
		// 更新订单备注（用于 checkout 时用户修改已存在 pending 订单的商家备注）
		auth.POST("/user/order/updateNotes", controller.UpdateOrderNotes)
		auth.POST("/user/order/updateAddress", controller.UpdateOrderAddress)

		// User address management
		auth.GET("/user/addresses", controller.GetUserAddresses)
		auth.POST("/user/address", controller.AddUserAddress)
		auth.PUT("/user/address/:id", controller.EditUserAddress)
		auth.POST("/user/address/:id/default", controller.SetDefaultAddress)
		auth.DELETE("/user/address/:id", controller.DeleteUserAddress)

	}
	return fe
}
