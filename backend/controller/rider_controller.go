package controller

import (
	"backend/global"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// GET /rider/info
func GetRiderInfo(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID") // 从 AuthMiddleware 取用户ID

	var profile models.RiderProfile
	err := global.Db.Where("user_id = ?", baseUserID).First(&profile).Error
	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "未找到骑手信息"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": profile})
}

// POST /rider/status
func UpdateRiderStatus(c *gin.Context) {
	var req struct {
		IsOnline bool `json:"isOnline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	baseUserID := c.GetUint("baseUserID")

	err := global.Db.Model(&models.RiderProfile{}).Where("user_id = ?", baseUserID).
		Update("is_online", req.IsOnline).Error

	if err != nil {
		c.JSON(200, gin.H{"code": 0, "msg": "更新失败"})
		return
	}

	c.JSON(200, gin.H{"code": 1, "data": gin.H{"success": true}})
}
