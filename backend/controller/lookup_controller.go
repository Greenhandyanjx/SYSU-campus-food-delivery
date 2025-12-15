package controller

import (
	"backend/global"
	"backend/models"
	butils "backend/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetMerchantDetail 返回商家基础信息，参数: id 或 base_id
func GetMerchantDetail(c *gin.Context) {
	idStr := c.Query("id")
	baseIDStr := c.Query("base_id")
	var m models.Merchant
	var err error
	if idStr != "" {
		var id uint64
		id, err = strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid id"})
			return
		}
		// 先尝试从缓存读取
		cacheKey := "merchant:id:" + strconv.FormatUint(id, 10)
		if v, ok := butils.DefaultCache.Get(cacheKey); ok {
			if mm, ok2 := v.(models.Merchant); ok2 {
				m = mm
			}
		} else {
			if err = global.Db.First(&m, uint(id)).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found"})
				return
			}
			butils.DefaultCache.Set(cacheKey, m, 5*time.Minute)
		}
	} else if baseIDStr != "" {
		var baseID uint64
		baseID, err = strconv.ParseUint(baseIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid base_id"})
			return
		}
		cacheKey := "merchant:base_id:" + strconv.FormatUint(baseID, 10)
		if v, ok := butils.DefaultCache.Get(cacheKey); ok {
			if mm, ok2 := v.(models.Merchant); ok2 {
				m = mm
			}
		} else {
			if err = global.Db.Where("base_id = ?", uint(baseID)).First(&m).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found"})
				return
			}
			butils.DefaultCache.Set(cacheKey, m, 5*time.Minute)
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "id or base_id required"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": m})
}

// GetMerchantProfile 返回当前认证商家的信息（从中间件读取 baseUserID）
func GetMerchantProfile(c *gin.Context) {
	baseIDIf, ok := c.Get("baseUserID")
	if !ok {
		c.JSON(401, gin.H{"code": 0, "msg": "no user in context"})
		return
	}
	var baseID uint
	switch v := baseIDIf.(type) {
	case uint:
		baseID = v
	case int:
		baseID = uint(v)
	case int64:
		baseID = uint(v)
	case float64:
		baseID = uint(v)
	default:
		c.JSON(500, gin.H{"code": 0, "msg": "invalid user id type"})
		return
	}
	var m models.Merchant
	if err := global.Db.Where("base_id = ?", baseID).First(&m).Error; err != nil {
		c.JSON(404, gin.H{"code": 0, "msg": "merchant not found"})
		return
	}
	c.JSON(200, gin.H{"code": 1, "data": m})
}

// UpdateMerchantProfile 用于认证商家更新自己的信息（shop name / phone / logo / shop_location / owner）
func UpdateMerchantProfile(c *gin.Context) {
	baseIDIf, ok := c.Get("baseUserID")
	if !ok {
		c.JSON(401, gin.H{"code": 0, "msg": "no user in context"})
		return
	}
	var baseID uint
	switch v := baseIDIf.(type) {
	case uint:
		baseID = v
	case int:
		baseID = uint(v)
	case int64:
		baseID = uint(v)
	case float64:
		baseID = uint(v)
	default:
		c.JSON(500, gin.H{"code": 0, "msg": "invalid user id type"})
		return
	}

	var payload struct {
		ShopName     string `json:"shop_name"`
		Phone        string `json:"phone"`
		Logo         string `json:"logo"`
		ShopLocation string `json:"shop_location"`
		Owner        string `json:"owner"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(400, gin.H{"code": 0, "msg": "invalid body"})
		return
	}

	updates := make(map[string]interface{})
	if payload.ShopName != "" {
		updates["shop_name"] = payload.ShopName
	}
	if payload.Phone != "" {
		updates["phone"] = payload.Phone
	}
	if payload.Logo != "" {
		updates["logo"] = payload.Logo
	}
	if payload.ShopLocation != "" {
		updates["shop_location"] = payload.ShopLocation
	}
	if payload.Owner != "" {
		updates["owner"] = payload.Owner
	}
	if len(updates) == 0 {
		c.JSON(200, gin.H{"code": 1, "msg": "no changes"})
		return
	}
	if err := global.Db.Model(&models.Merchant{}).Where("base_id = ?", baseID).Updates(updates).Error; err != nil {
		c.JSON(500, gin.H{"code": 0, "msg": "update failed"})
		return
	}
	c.JSON(200, gin.H{"code": 1, "msg": "updated"})
}

// GetBaseUserDetail 返回 base_users 表的基本信息；如果未提供 id，则从 Authorization token 推断
func GetBaseUserDetail(c *gin.Context) {
	idStr := c.Query("id")
	var b models.BaseUser
	var err error
	if idStr == "" {
		// 尝试从 Authorization 头解析 token
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "id or Authorization required"})
			return
		}
		username, err := butils.ParseJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "invalid token"})
			return
		}
		// 尝试缓存
		cacheKey := "baseuser:username:" + username
		if v, ok := butils.DefaultCache.Get(cacheKey); ok {
			if bb, ok2 := v.(models.BaseUser); ok2 {
				b = bb
			}
		} else {
			if err = global.Db.Where("username = ?", username).First(&b).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "user not found"})
				return
			}
			butils.DefaultCache.Set(cacheKey, b, 5*time.Minute)
		}
	} else {
		var id uint64
		id, err = strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid id"})
			return
		}
		cacheKey := "baseuser:id:" + strconv.FormatUint(id, 10)
		if v, ok := butils.DefaultCache.Get(cacheKey); ok {
			if bb, ok2 := v.(models.BaseUser); ok2 {
				b = bb
			}
		} else {
			if err = global.Db.First(&b, uint(id)).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "user not found"})
				return
			}
			butils.DefaultCache.Set(cacheKey, b, 5*time.Minute)
		}
	}
	// 避免返回密码字段
	resp := map[string]interface{}{
		"id":       b.ID,
		"username": b.Username,
		"role":     b.Role,
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": resp})
}

// 注意：为了进一步优化，建议为 base_users.username 和 merchants.base_id 创建索引。
// 你可以将以下 SQL 作为 migration：
// ALTER TABLE base_users ADD INDEX idx_base_users_username (username);
// ALTER TABLE merchants ADD INDEX idx_merchants_base_id (base_id);
