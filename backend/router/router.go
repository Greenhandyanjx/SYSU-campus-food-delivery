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
		auth.POST("/register", controller.Register)
	}
	return fe
}