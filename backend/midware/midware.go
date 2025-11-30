package midware

import (
	"log"
	"net/http"
	"strings"

	"backend/global"
	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuthMiddleware 提取请求头中的 authorization token 并解析它
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 放行不需要鉴权的路径（例如注册页的图片上传）
		if c.Request.Method == "OPTIONS" || c.Request.URL.Path == "/api/common/upload" || c.Request.URL.Path == "/api/sms/send" || c.Request.URL.Path == "/api/sms/verify" {
			c.Next()
			return
		}
		// 提取 authorization 头或 query token（支持 websocket 使用 query token）
		authHeader := c.GetHeader("Authorization")
		tokenString := ""
		if authHeader != "" {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// 支持把 token 放在查询参数 token 中（供 WebSocket 握手或特殊客户端使用）
			tokenQuery := c.Query("token")
			if tokenQuery != "" {
				tokenString = tokenQuery
			}
		}

		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": "401",
				"msg":  "no token provided",
			})
			return
		}

		// 解析 token
		username, err := utils.ParseJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code": "401",
				"msg":  "Invalid token"})
			return
		}
		// 根据用户名查找用户ID
		var baseUser models.BaseUser
		if err := global.Db.Where("username = ?", username).First(&baseUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code": "401",
					"msg":  "用户未找到",
				})
			} else {
				log.Printf("数据库查询错误: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": "500",
					"msg":  "服务器内部错误，请稍后再试",
				})
			}
			c.Abort()
			return
		}
		// 将用户ID存入上下文
		c.Set("baseUserID", baseUser.ID)
		c.Next()
	}
}
