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
	UserBaseID     uint      `json:"user_base_id"`
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

	// 优化实现：一次性聚合查询，获取每个 merchant 的最后一条消息、时间、未读计数，并 Join 商家信息
	// 同时在运行时尝试创建加速索引（仅当数据库为 MySQL 且索引不存在时）

	ensureIndexes := func() {
		var cnt int64
		global.Db.Raw(`SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?`, "chat_messages", "idx_chat_user_merchant_created_at").Scan(&cnt)
		if cnt == 0 {
			global.Db.Exec(`ALTER TABLE chat_messages ADD INDEX idx_chat_user_merchant_created_at (user_base_id, merchant_id, created_at)`)
		}
		global.Db.Raw(`SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?`, "chat_messages", "idx_chat_user_merchant_from_status").Scan(&cnt)
		if cnt == 0 {
			global.Db.Exec(`ALTER TABLE chat_messages ADD INDEX idx_chat_user_merchant_from_status (user_base_id, merchant_id, from_base_id, status)`)
		}
	}

	ensureIndexes()

	type row struct {
		MerchantID     uint      `json:"merchant_id"`
		LastMessage    string    `json:"last_message"`
		LastAt         time.Time `json:"last_at"`
		UnreadCount    int64     `json:"unread_count"`
		UserBaseID     uint      `json:"user_base_id"`
		MerchantName   string    `json:"merchant_name"`
		MerchantAvatar string    `json:"merchant_avatar"`
	}

	// 聚合查询：先取每个 merchant 的最后消息，再统计未读（来自商家且 status != 'read'），最后 Join 商家信息
	sql := `
		SELECT t.merchant_id, t.last_message, t.last_at, t.user_base_id, COALESCE(u.unread_count,0) AS unread_count, mm.shop_name AS merchant_name, mm.logo AS merchant_avatar
		FROM (
			SELECT cm.merchant_id, cm.content AS last_message, cm.created_at AS last_at, cm.user_base_id
			FROM chat_messages cm
			JOIN (
				SELECT merchant_id, MAX(created_at) AS last_at
				FROM chat_messages
				WHERE user_base_id = ?
				GROUP BY merchant_id
			) s ON cm.merchant_id = s.merchant_id AND cm.created_at = s.last_at
			WHERE cm.user_base_id = ?
		) t
		LEFT JOIN (
			SELECT cm.merchant_id, COUNT(*) AS unread_count
			FROM chat_messages cm
			JOIN merchants m ON m.id = cm.merchant_id
			WHERE cm.user_base_id = ? AND cm.from_base_id = m.base_id AND cm.status != 'read'
			GROUP BY cm.merchant_id
		) u ON t.merchant_id = u.merchant_id
		LEFT JOIN merchants mm ON mm.id = t.merchant_id
		ORDER BY t.last_at DESC
		`

	var rows []row
	if err := global.Db.Raw(sql, userBaseID, userBaseID, userBaseID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db error", "error": err.Error()})
		return
	}

	summaries := make([]UserChatSummary, 0, len(rows))
	for _, r := range rows {
		summaries = append(summaries, UserChatSummary{
			MerchantID:     r.MerchantID,
			LastMessage:    r.LastMessage,
			LastAt:         r.LastAt,
			UnreadCount:    r.UnreadCount,
			UserBaseID:     r.UserBaseID,
			MerchantName:   r.MerchantName,
			MerchantAvatar: r.MerchantAvatar,
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

//111111
