package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// 连接管理
var (
	upgrader = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	// key: base_user_id (uint as string) -> conn
	connStore = make(map[uint]*websocket.Conn)
	connMu    sync.RWMutex
)

// ChatMessagePayload 用于 WS 收发的消息结构
type ChatMessagePayload struct {
	MerchantID   uint   `json:"merchant_id"`
	ToUserBaseID uint   `json:"to_user_base_id"` // 可选，当 merchant 作为发送者时指定
	Content      string `json:"content"`
	Type         string `json:"type"` // text/image
}

// ChatWS 处理 websocket 连接
func ChatWS(c *gin.Context) {
	// 握手鉴权：优先使用 token；若无 token 则尝试使用 query uid（仅作开发/兼容）
	token := c.Query("token")
	if token == "" {
		token = c.GetHeader("Authorization")
	}

	var base models.BaseUser
	if token != "" {
		username, err := utils.ParseJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "invalid token"})
			return
		}
		if err := global.Db.Where("username = ?", username).First(&base).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "user not found"})
			return
		}
	} else {
		// 尝试从 query 参数读取 uid（不推荐用于生产）
		uidStr := c.Query("uid")
		if uidStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "no token or uid provided"})
			return
		}
		// 把 uid 转为 uint 并查询 base user
		var uid uint64
		uid, _ = strconv.ParseUint(uidStr, 10, 64)
		if uid == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid uid"})
			return
		}
		if err := global.Db.First(&base, uid).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "user not found"})
			return
		}
	}

	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	// 注册连接
	connMu.Lock()
	connStore[base.ID] = ws
	connMu.Unlock()

	// 简单的读循环：接收消息并处理
	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		var payload ChatMessagePayload
		if err := json.Unmarshal(message, &payload); err != nil {
			// 忽略错误消息
			continue
		}

		// 保存消息到 DB
		chat := models.ChatMessage{
			FromBaseID: base.ID,
			MerchantID: payload.MerchantID,
			Content:    payload.Content,
			Type:       payload.Type,
			Status:     "sent",
			CreatedAt:  time.Now(),
		}

		// 如果发送者是 merchant，则 userBaseId 需要从 payload.ToUserBaseID 获取
		// 如果发送者是 user，则 userBaseId = base.ID
		// 尝试检查发送者角色
		var merchant models.Merchant
		if err := global.Db.Where("base_id = ?", base.ID).First(&merchant).Error; err == nil {
			// 发送者是商家
			chat.UserBaseID = payload.ToUserBaseID
		} else {
			// 发送者被视为用户
			chat.UserBaseID = base.ID
		}

		if err := global.Db.Create(&chat).Error; err != nil {
			// 存储失败，忽略
		}

		// 发送到接收方（商家或用户）如果在线
		// 对于用户发送者：接收方是商家 — 我们需要把消息推给商家端（商家连接存为其 base_id）
		// 对于商家发送者：接收方是用户（payload.ToUserBaseID）

		// 查找目标 base_id
		var targetBaseID uint
		if chat.UserBaseID == base.ID {
			// 发送者是用户，目标为商家对应的 base_id
			var targetMerchant models.Merchant
			if err := global.Db.First(&targetMerchant, chat.MerchantID).Error; err == nil {
				targetBaseID = targetMerchant.BaseID
			}
		} else {
			// 发送者是商家，目标为用户的 base_id
			targetBaseID = chat.UserBaseID
		}

		if targetBaseID != 0 {
			connMu.RLock()
			targetConn, ok := connStore[targetBaseID]
			connMu.RUnlock()
			if ok && targetConn != nil {
				// 转发原始消息（可扩展为带时间戳、id 等）
				out := map[string]interface{}{
					"from_base_id": chat.FromBaseID,
					"merchant_id":  chat.MerchantID,
					"user_base_id": chat.UserBaseID,
					"content":      chat.Content,
					"type":         chat.Type,
					"created_at":   chat.CreatedAt,
				}
				_ = targetConn.WriteJSON(out)
				// 更新状态为 delivered
				now := time.Now()
				chat.Status = "delivered"
				chat.DeliveredAt = &now
				global.Db.Model(&models.ChatMessage{}).Where("id = ?", chat.ID).Updates(map[string]interface{}{"status": chat.Status, "delivered_at": chat.DeliveredAt})
			}
		}
	}

	// 连接关闭，移除
	connMu.Lock()
	delete(connStore, base.ID)
	connMu.Unlock()
	ws.Close()
}

// ChatHistory 返回指定商家与用户之间的历史消息，分页
func ChatHistory(c *gin.Context) {
	merchantId := c.Query("merchantId")
	userBaseId := c.Query("userBaseId")
	if merchantId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "merchantId required"})
		return
	}

	// 若未提供 userBaseId，则尝试从 Authorization token 推断
	if userBaseId == "" {
		token := c.GetHeader("Authorization")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "userBaseId required or provide Authorization token"})
			return
		}
		username, err := utils.ParseJWT(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "invalid token"})
			return
		}
		var base models.BaseUser
		if err := global.Db.Where("username = ?", username).First(&base).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 0, "msg": "user not found"})
			return
		}
		userBaseId = strconv.FormatUint(uint64(base.ID), 10)
	}

	var msgs []models.ChatMessage
	// 简单分页参数
	page := 1
	pageSize := 50
	if err := global.Db.Where("merchant_id = ? AND user_base_id = ?", merchantId, userBaseId).
		Order("created_at desc").Limit(pageSize).Offset((page - 1) * pageSize).Find(&msgs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "db error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": msgs})
}
