package main

import (
	"backend/config"
	"backend/router"
	"fmt"
	"log"
)

func main() {
	// 1. 加载配置
	config.InitConfig()

	// 2. 初始化数据库连接 ⭐【新增/恢复】——不要再依赖 InitConfig 里偷偷初始化
	config.InitDB()
	// global.Db.AutoMigrate(&models.PayInfo{})
	// 3. 执行自动建表（AutoMigrate 是幂等的，可以每次启动都跑，不会清空数据）
	// if err := config.Initalldb(); err != nil { // ⭐【改成带错误判断】
	// 	log.Fatalf("Initalldb failed: %v", err)
	// }

	// 4. 初始化路由
	r := router.SetRouter()

	// 5. 启动服务
	if err := r.Run(":" + config.AppConfig.App.Port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}

	fmt.Println("server start at port", config.AppConfig.App.Port) // ⭐ 顺带改个真实端口
}
