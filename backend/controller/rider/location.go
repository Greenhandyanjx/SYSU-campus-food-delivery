package rider

import (
	"backend/global"
	"backend/models"

	"github.com/gin-gonic/gin"
)

func UpdateLocation(c *gin.Context) {
	baseID := c.GetUint("baseUserID")
	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Address   string  `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误")
		return
	}

	if err := global.Db.Model(&models.RiderProfile{}).
		Where("user_id = ?", baseID).
		Updates(map[string]any{
			"latitude":  req.Latitude,
			"longitude": req.Longitude,
			"address":   req.Address,
		}).Error; err != nil {
		fail(c, "更新失败")
		return
	}
	ok(c, gin.H{"success": true})
}
