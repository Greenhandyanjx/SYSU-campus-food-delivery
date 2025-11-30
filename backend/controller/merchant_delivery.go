package controller

import (
	"backend/global"
	"backend/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetMerchantDeliveryConfig 返回指定商家对应的配送配置。
// 参数支持两种：`base_id`（商家 base_id）或 `id`（商家主键 id）。
func GetMerchantDeliveryConfig(c *gin.Context) {
	// 尝试先读取 id（优先），再回退到 base_id
	idStr := c.Query("id")
	baseIDStr := c.Query("base_id")

	var baseID uint64

	if idStr != "" {
		// 前端传了商家主键 id，先根据 id 查出对应 merchant 的 base_id
		mid, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid id"})
			return
		}
		var m models.Merchant
		if err := global.Db.Where("id = ?", uint(mid)).First(&m).Error; err != nil {
			// 未找到 merchant，则回退到使用传入 id 作为 base_id 尝试读取配置
			baseID = mid
		} else {
			baseID = uint64(m.BaseID)
		}
	} else if baseIDStr != "" {
		bid, err := strconv.ParseUint(baseIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid base_id"})
			return
		}
		baseID = bid
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "missing id or base_id"})
		return
	}

	var cfg models.MerchantDeliveryConfig
	if err := global.Db.Where("base_id = ?", uint(baseID)).First(&cfg).Error; err != nil {
		// 未找到则返回默认值（但不写入）
		defaultCfg := models.MerchantDeliveryConfig{
			BaseID:        uint(baseID),
			MinPrice:      15,
			DeliveryFee:   2,
			DeliveryRange: 2000,
		}
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": defaultCfg})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": cfg})
}
