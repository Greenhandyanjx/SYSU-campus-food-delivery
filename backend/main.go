package main

import (
	"backend/config"
	// "backend/global"
	// "backend/models"
	"backend/router"
	"backend/utils"
	"fmt"
	"log"
)

func main() {
	// 1. 加载配置
	config.InitConfig()

	// 2. 初始化数据库连接并执行自动建表（AutoMigrate 是幂等的，可以每次启动都跑）
	// config.InitDB()
	// if err := global.Db.AutoMigrate(&models.User{}); err != nil {
	// 	log.Fatalf("AutoMigrate user failed: %v", err)
	// }
	// // 3. 执行自动建表（AutoMigrate 是幂等的，可以每次启动都跑，不会清空数据）
	// if err := config.Initalldb(); err != nil {
	// 	log.Fatalf("Initalldb failed: %v", err)
	// }

	// 启动后台清理 goroutine：定期清除过期的 pending 订单及其关联的 order_dishes/order_meals
	// 启动后台清理 goroutine
	go utils.StartPendingCleanup()

	// 4. 初始化路由
	r := router.SetRouter()

	// 5. 启动服务
	if err := r.Run(":" + config.AppConfig.App.Port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}

	fmt.Println("server start at port", config.AppConfig.App.Port) // ⭐ 顺带改个真实端口
}
