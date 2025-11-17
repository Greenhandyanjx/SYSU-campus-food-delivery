package controller

import (
	"backend/global"
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// SendCode 生成验证码并存入 Redis，返回验证码以便在注册页面显示（开发用）
func SendCode(c *gin.Context) {
	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Purpose string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "0", "msg": "phone required"})
		return
	}
	if req.Purpose == "" {
		req.Purpose = "register"
	}
	// 生成 6 位数字验证码
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	// 尝试将验证码写入 Redis
	if global.RedisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		key := fmt.Sprintf("sms:%s:%s", req.Purpose, req.Phone)
		// 存储 5 分钟
		_ = global.RedisClient.Set(ctx, key, code, 5*time.Minute).Err()
	}

	// 返回验证码（开发环境可见）后面我们注册阿里云短信服务再改这里
	c.JSON(http.StatusOK, gin.H{"code": "1", "msg": "ok", "data": gin.H{"code": code}})
}

// VerifyCode 校验验证码
func VerifyCode(c *gin.Context) {
	var req struct {
		Phone   string `json:"phone" binding:"required"`
		Code    string `json:"code" binding:"required"`
		Purpose string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": "0", "msg": "params required"})
		return
	}
	if req.Purpose == "" {
		req.Purpose = "register"
	}

	if global.RedisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		key := fmt.Sprintf("sms:%s:%s", req.Purpose, req.Phone)
		v, err := global.RedisClient.Get(ctx, key).Result()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "0", "msg": "code invalid or expired"})
			return
		}
		if v != req.Code {
			c.JSON(http.StatusUnauthorized, gin.H{"code": "0", "msg": "code mismatch"})
			return
		}
		// 校验通过：删除 key
		_ = global.RedisClient.Del(ctx, key).Err()
		c.JSON(http.StatusOK, gin.H{"code": "1", "msg": "ok"})
		return
	}
	// 如果没有 Redis，则在无状态模式下拒绝
	c.JSON(http.StatusInternalServerError, gin.H{"code": "0", "msg": "redis not enabled"})
}
