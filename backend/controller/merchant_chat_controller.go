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

	// 优化：一次性聚合查询，避免 N+1。思路：
	// 1) 找出每个 user_base_id 的最后一条消息
	// 2) 统计每个 user_base_id 的未读数（来自用户且 status != 'read'）
	// 3) 将两部分 join 并按最后时间排序返回

	// 在数据库中创建必要的索引以加速查询（如果尚未存在）
	// 这些操作做存在性检查，避免重复创建导致错误
	ensureIndexes := func() {
		// 使用 information_schema 检查索引是否存在（MySQL）
		var cnt int64
		// 索引：merchant_id, user_base_id, created_at
		global.Db.Raw(`SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?`, "chat_messages", "idx_chat_merchant_user_created_at").Scan(&cnt)
		if cnt == 0 {
			global.Db.Exec(`ALTER TABLE chat_messages ADD INDEX idx_chat_merchant_user_created_at (merchant_id, user_base_id, created_at)`)
		}
		// 索引：merchant_id, user_base_id, from_base_id, status
		global.Db.Raw(`SELECT COUNT(1) FROM INFORMATION_SCHEMA.STATISTICS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND INDEX_NAME = ?`, "chat_messages", "idx_chat_merchant_user_from_status").Scan(&cnt)
		if cnt == 0 {
			global.Db.Exec(`ALTER TABLE chat_messages ADD INDEX idx_chat_merchant_user_from_status (merchant_id, user_base_id, from_base_id, status)`)
		}
	}

	// 尝试创建索引（若数据库为非 MySQL，该操作会静默失败）
	ensureIndexes()

	// 执行聚合查询（一次性获取最后一条消息与未读计数）
	type row struct {
		UserBaseID  uint      `json:"user_base_id"`
		LastMessage string    `json:"last_message"`
		LastAt      time.Time `json:"last_at"`
		UnreadCount int64     `json:"unread_count"`
	}

	sql := `
	SELECT t.user_base_id, t.last_message, t.last_at, COALESCE(u.unread_count,0) AS unread_count
	FROM (
	  SELECT cm.user_base_id, cm.content AS last_message, cm.created_at AS last_at
	  FROM chat_messages cm
	  JOIN (
		SELECT user_base_id, MAX(created_at) AS last_at
		FROM chat_messages
		WHERE merchant_id = ?
		GROUP BY user_base_id
	  ) s ON cm.user_base_id = s.user_base_id AND cm.created_at = s.last_at
	  WHERE cm.merchant_id = ?
	) t
	LEFT JOIN (
	  SELECT user_base_id, COUNT(*) AS unread_count
	  FROM chat_messages
	  WHERE merchant_id = ? AND from_base_id = user_base_id AND status != 'read'
	  GROUP BY user_base_id
	) u ON t.user_base_id = u.user_base_id
	ORDER BY t.last_at DESC
	`

	var rows []row
	if err := global.Db.Raw(sql, merchantID, merchantID, merchantID).Scan(&rows).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db error", "error": err.Error()})
		return
	}

	summaries := make([]MerchantChatSummary, 0, len(rows))
	for _, r := range rows {
		summaries = append(summaries, MerchantChatSummary{
			MerchantID:  uint(merchantID),
			UserBaseID:  r.UserBaseID,
			LastMessage: r.LastMessage,
			LastAt:      r.LastAt,
			UnreadCount: r.UnreadCount,
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

//前有无意义
