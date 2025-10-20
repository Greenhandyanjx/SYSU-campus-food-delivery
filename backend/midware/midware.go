package midware

import (
	"net/http"
	"strings"

	"backend/utils"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware 提取请求头中的 authorization token 并解析它
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 提取 authorization 头
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":"401",
				"msg": "no token provided",
			})
			return
		}

		// 提取 token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":"401",
				"msg": "Invalid Authorization header format",
			})
			return
		}

		// 解析 token
		username, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
	            "code":"401",		
				"msg": "Invalid token"})
			return
		}

		// 将解析后的用户名存储在上下文中
		c.Set("username", username)

		// 调用下一个处理程序
		c.Next()
	}
}
