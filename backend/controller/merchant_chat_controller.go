package controller

import (
	"backend/global"
	"backend/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MerchantChatSummary 表示商家侧的会话摘要
type MerchantChatSummary struct {
	MerchantID  uint      `json:"merchant_id"`
	UserBaseID  uint      `json:"user_base_id"`
	LastMessage string    `json:"last_message"`
	LastAt      time.Time `json:"last_at"`
	UnreadCount int64     `json:"unread_count"`
}

// GetMerchantChats 返回指定商家的会话列表（若不传 merchantId，则从 token 推断商家）
func GetMerchantChats(c *gin.Context) {
	merchantIdStr := c.Query("merchantId")
	var merchantID uint64
	var err error

	if merchantIdStr != "" {
		merchantID, err = strconv.ParseUint(merchantIdStr, 10, 64)
		if err != nil || merchantID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid merchantId"})
			return
		}
	} else {
		// 从中间件注入的 baseUserID 推断商家
		v, ok := c.Get("baseUserID")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "merchantId required or authenticated merchant"})
			return
		}
		baseID := v.(uint)
		var m models.Merchant
		if err := global.Db.Where("base_id = ?", baseID).First(&m).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found for current user"})
			return
		}
		merchantID = uint64(m.ID)
	}

	// 获取参与会话的 user_base_id 列表
	var userIDs []uint
	if err := global.Db.Model(&models.ChatMessage{}).Distinct("user_base_id").Where("merchant_id = ?", merchantID).Pluck("user_base_id", &userIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db error"})
		return
	}

	summaries := make([]MerchantChatSummary, 0, len(userIDs))
	for _, uid := range userIDs {
		var last models.ChatMessage
		if err := global.Db.Where("merchant_id = ? AND user_base_id = ?", merchantID, uid).Order("created_at desc").Limit(1).Find(&last).Error; err != nil {
			continue
		}
		var unread int64
		// unread for merchant means messages sent by user (from_base_id == user_base_id) and not read
		global.Db.Model(&models.ChatMessage{}).Where("merchant_id = ? AND user_base_id = ? AND from_base_id = ? AND status != ?", merchantID, uid, uid, "read").Count(&unread)

		summaries = append(summaries, MerchantChatSummary{
			MerchantID:  uint(merchantID),
			UserBaseID:  uid,
			LastMessage: last.Content,
			LastAt:      last.CreatedAt,
			UnreadCount: unread,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": summaries})
}

// MarkChatRead 标记商家与指定用户的会话为已读（将用户发来的消息标记为 read）
func MarkChatRead(c *gin.Context) {
	var payload struct {
		MerchantID uint `json:"merchant_id"`  // 可选
		UserBaseID uint `json:"user_base_id"` // 必填
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid payload"})
		return
	}
	if payload.UserBaseID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "user_base_id required"})
		return
	}

	// 如果未提供 merchant_id，则从上下文中推断
	var merchantID uint
	if payload.MerchantID != 0 {
		merchantID = payload.MerchantID
	} else {
		v, ok := c.Get("baseUserID")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "merchant_id required or authenticated merchant"})
			return
		}
		baseID := v.(uint)
		var m models.Merchant
		if err := global.Db.Where("base_id = ?", baseID).First(&m).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found for current user"})
			return
		}
		merchantID = m.ID
	}

	now := time.Now()
	// 将用户发出的消息（from_base_id == user_base_id）标记为 read
	if err := global.Db.Model(&models.ChatMessage{}).Where("merchant_id = ? AND user_base_id = ? AND from_base_id = ? AND status != ?", merchantID, payload.UserBaseID, payload.UserBaseID, "read").Updates(map[string]interface{}{"status": "read", "read_at": &now}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db update error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "ok"})
}
