package main

import (
	"backend/config"
	"backend/router"
	"fmt"
	"log"
)	

func main() {
    config.InitConfig()
	r := router.SetRouter() // 初始化路由
    if err := r.Run(":" + config.AppConfig.App.Port); err != nil {
        log.Fatalf("Server startup failed: %v", err)
    }
	fmt.Println("server start at port 3000")
}