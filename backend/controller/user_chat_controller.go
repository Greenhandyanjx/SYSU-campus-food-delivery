package controller

import (
	"backend/global"
	"backend/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UserChatSummary 表示用户侧的会话摘要
type UserChatSummary struct {
	MerchantID     uint      `json:"merchant_id"`
	LastMessage    string    `json:"last_message"`
	LastAt         time.Time `json:"last_at"`
	UnreadCount    int64     `json:"unread_count"`
	MerchantName   string    `json:"merchant_name"`
	MerchantAvatar string    `json:"merchant_avatar"`
}

// GetUserChats 返回当前用户的会话列表（包含商家信息、最后一条消息与未读数）
func GetUserChats(c *gin.Context) {
	v, ok := c.Get("baseUserID")
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "unauthenticated"})
		return
	}
	userBaseID := v.(uint)

	// 查询与当前用户有过会话的 merchant_id 列表
	var merchantIDs []uint
	if err := global.Db.Model(&models.ChatMessage{}).Distinct("merchant_id").Where("user_base_id = ?", userBaseID).Pluck("merchant_id", &merchantIDs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db error"})
		return
	}

	summaries := make([]UserChatSummary, 0, len(merchantIDs))
	for _, mid := range merchantIDs {
		var last models.ChatMessage
		if err := global.Db.Where("merchant_id = ? AND user_base_id = ?", mid, userBaseID).Order("created_at desc").Limit(1).Find(&last).Error; err != nil {
			continue
		}

		var m models.Merchant
		if err := global.Db.First(&m, mid).Error; err != nil {
			// 若找不到商家信息，仍然返回会话但填空商家字段
			summaries = append(summaries, UserChatSummary{
				MerchantID:  mid,
				LastMessage: last.Content,
				LastAt:      last.CreatedAt,
			})
			continue
		}

		var unread int64
		// 统计来自商家的未读消息（from_base_id == merchant.base_id）
		global.Db.Model(&models.ChatMessage{}).Where("merchant_id = ? AND user_base_id = ? AND from_base_id = ? AND status != ?", mid, userBaseID, m.BaseID, "read").Count(&unread)

		summaries = append(summaries, UserChatSummary{
			MerchantID:     mid,
			LastMessage:    last.Content,
			LastAt:         last.CreatedAt,
			UnreadCount:    unread,
			MerchantName:   m.ShopName,
			MerchantAvatar: m.Logo,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": summaries})
}

// MarkUserChatRead 标记用户与指定商家的会话为已读（将商家发来的消息标记为 read）
func MarkUserChatRead(c *gin.Context) {
	var payload struct {
		MerchantID uint `json:"merchant_id"`
		UserBaseID uint `json:"user_base_id"` // 可选，通常从 token 推断
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid payload"})
		return
	}

	// 推断 userBaseID
	var userBaseID uint
	if payload.UserBaseID != 0 {
		userBaseID = payload.UserBaseID
	} else {
		v, ok := c.Get("baseUserID")
		if !ok {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "user not authenticated"})
			return
		}
		userBaseID = v.(uint)
	}

	if payload.MerchantID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "merchant_id required"})
		return
	}

	// 查找商家的 base_id 用于比对 from_base_id
	var m models.Merchant
	if err := global.Db.First(&m, payload.MerchantID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found"})
		return
	}

	now := time.Now()
	if err := global.Db.Model(&models.ChatMessage{}).Where("merchant_id = ? AND user_base_id = ? AND from_base_id = ? AND status != ?", payload.MerchantID, userBaseID, m.BaseID, "read").Updates(map[string]interface{}{"status": "read", "read_at": &now}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db update error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "msg": "ok"})
}
