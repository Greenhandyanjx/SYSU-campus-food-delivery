package rider

import (
	"backend/global"
	"backend/models"

	"github.com/gin-gonic/gin"
)

// GET /api/rider/me
func GetMe(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")

	var p models.RiderProfile
	if err := global.Db.Where("user_id = ?", baseUserID).First(&p).Error; err != nil {
		fail(c, "未找到骑手信息")
		return
	}

	ok(c, gin.H{
		"id":              p.UserID,
		"name":            p.Name,
		"avatar":          p.Avatar,
		"phone":           p.Phone,
		"rating":          p.Rating,
		"completedOrders": p.CompletedOrders,
		"isOnline":        p.IsOnline,
	})
}

// PATCH /api/rider/me/online
func UpdateOnline(c *gin.Context) {
	var req struct {
		IsOnline bool `json:"isOnline"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误")
		return
	}

	baseUserID := c.GetUint("baseUserID")
	if err := global.Db.Model(&models.RiderProfile{}).
		Where("user_id = ?", baseUserID).
		Update("is_online", req.IsOnline).Error; err != nil {
		fail(c, "更新失败")
		return
	}

	ok(c, gin.H{"success": true})
}
