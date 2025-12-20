package rider

import (
	"backend/global"
	"backend/models"
	"errors"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// POST /api/rider/location
// body: { "latitude": 22.3, "longitude": 113.5, "address": "xxx(可选)" }
func UpdateLocation(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")
	if baseUserID == 0 {
		fail(c, "未登录")
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Address   string  `json:"address"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		fail(c, "参数错误")
		return
	}

	// 简单合法性校验（防止前端传 0/0 或 NaN）
	if req.Latitude < -90 || req.Latitude > 90 || req.Longitude < -180 || req.Longitude > 180 {
		fail(c, "经纬度不合法")
		return
	}

	updates := map[string]any{
		"latitude":  req.Latitude,
		"longitude": req.Longitude,
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}

	// 先更新（大多数情况 profile 已存在）
	tx := global.Db.Model(&models.RiderProfile{}).
		Where("user_id = ?", baseUserID).
		Updates(updates)

	if tx.Error != nil {
		fail(c, "更新定位失败")
		return
	}

	// 如果没更新到（说明没 profile），就创建一个最小 profile 再写一次
	if tx.RowsAffected == 0 {
		p := models.RiderProfile{
			UserID:      baseUserID,
			Latitude:    req.Latitude,
			Longitude:   req.Longitude,
			Address:     req.Address,
			IsOnline:    false,
			Name:        "",
			Avatar:      "",
			Phone:       "",
			Rating:      0,
			OnlineHours: 0,
		}
		if err := global.Db.Create(&p).Error; err != nil {
			// 并发情况下可能刚好被别人创建，遇到重复就再更新一次
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				_ = global.Db.Model(&models.RiderProfile{}).
					Where("user_id = ?", baseUserID).
					Updates(updates).Error
				ok(c, gin.H{"success": true})
				return
			}
			fail(c, "创建骑手信息失败")
			return
		}
	}

	ok(c, gin.H{"success": true})
}
