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

	// 2. 初始化数据库连接
	// 这里注释掉了，因为已经在config.InitConfig()中初始化了
	// config.InitDB()

	// 3. 执行自动建表
<<<<<<< HEAD
	//config.Initalldb()

=======
	//不改动数据执行一次即可
	//config.Initalldb()
>>>>>>> dev
	// 4. 初始化路由
	r := router.SetRouter()

	// 5. 启动服务
	if err := r.Run(":" + config.AppConfig.App.Port); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}

	fmt.Println("server start at port 3000")
}
