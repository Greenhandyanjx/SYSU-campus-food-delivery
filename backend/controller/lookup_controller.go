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
