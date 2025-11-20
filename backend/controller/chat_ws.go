package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"encoding/json"
	"log"
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
	MerchantID uint `json:"merchant_id"`
	// 兼容：客户端可能发送 `user_base_id`（user 作为发送者或作为目标时）
	// 或者发送 `to_user_base_id`（商家端可能使用此名字）。两者之一可能为 0。
	UserBaseID   uint   `json:"user_base_id"`
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
		log.Println("❌ WS Upgrade failed:", err)
		return
	}
	log.Println("✔ WS connected: base_id =", base.ID)
	// 注册连接
	connMu.Lock()
	connStore[base.ID] = ws
	connMu.Unlock()
	log.Println("✔ Registered WS conn for base_id =", base.ID)
	// 简单的读循环：接收消息并处理
	for {
		_, message, err := ws.ReadMessage()
		log.Println("📩 Incoming WS message from base_id =", base.ID, "raw =", string(message))
		if err != nil {
			break
		}
		var payload ChatMessagePayload
		if err := json.Unmarshal(message, &payload); err != nil {
			// 忽略错误消息
			log.Println("❌ Unmarshal failed:", err)

			continue
		}
		// 兼容各种客户端字段名：优先使用 user_base_id，其次使用 to_user_base_id
		var effectiveUserBaseID uint
		if payload.UserBaseID != 0 {
			effectiveUserBaseID = payload.UserBaseID
		} else {
			effectiveUserBaseID = payload.ToUserBaseID
		}
		log.Println("➡ Parsed payload:", payload, "effectiveUserBaseID=", effectiveUserBaseID)
		// 保存消息到 DB：更稳健地解析 merchant_id / user_base_id
		log.Println("📝 Preparing to store message: payloadMerchant=", payload.MerchantID, "from_base_id=", base.ID)

		chat := models.ChatMessage{
			FromBaseID: base.ID,
			Content:    payload.Content,
			Type:       payload.Type,
			Status:     "sent",
			CreatedAt:  time.Now(),
		}

		// 检查发送者是否为商家（通过 base_id 关联）
		var senderMerchant models.Merchant
		senderIsMerchant := false
		if err := global.Db.Where("base_id = ?", base.ID).First(&senderMerchant).Error; err == nil {
			senderIsMerchant = true
			// 发送者是商家，确保 chat.MerchantID 为该商家的 id
			chat.MerchantID = senderMerchant.ID
			// user id 从 effectiveUserBaseID 取得（必须由前端提供）
			chat.UserBaseID = effectiveUserBaseID
		} else {
			// 发送者被视为用户：userBaseId = 发送者 base id
			chat.UserBaseID = base.ID
			// merchant id 需要从 payload.MerchantID 解析：支持两种情况：
			// 1) 前端传入的是商家主键 id（merchant.id）
			// 2) 前端错误地传入了商家对应的 base_id（merchant.base_id），作为回退我们按 base_id 查找
			if payload.MerchantID != 0 {
				var targetMerchant models.Merchant
				// 先按主键查找
				if err := global.Db.First(&targetMerchant, payload.MerchantID).Error; err == nil {
					chat.MerchantID = targetMerchant.ID
				} else {
					// 回退：尝试按 base_id 查找
					if err := global.Db.Where("base_id = ?", payload.MerchantID).First(&targetMerchant).Error; err == nil {
						chat.MerchantID = targetMerchant.ID
					}
				}
			}
		}

		// 若 merchant id 仍然为 0（无法解析），记录并继续（消息仍会被存储但无法转发）
		if chat.MerchantID == 0 {
			log.Println("⚠️ merchant id unresolved for message from base_id=", base.ID, "payloadMerchant=", payload.MerchantID)
		}

		if err := global.Db.Create(&chat).Error; err != nil {
			log.Println("❌ failed to persist chat message:", err)
		}

		// 发送到接收方（商家或用户）如果在线：
		// 对于用户发送者（senderIsMerchant == false）：接收方是商家对应的 base_id
		// 对于商家发送者：接收方是用户（chat.UserBaseID）

		// 查找目标 base_id
		var targetBaseID uint
		if !senderIsMerchant {
			// 发送者是用户，目标为商家对应的 base_id
			if chat.MerchantID != 0 {
				var targetMerchant models.Merchant
				// 先按 merchant.id 查找
				if err := global.Db.First(&targetMerchant, chat.MerchantID).Error; err == nil {
					targetBaseID = targetMerchant.BaseID
				} else {
					// 回退：merchant.MerchantID 可能本身是 base_id 的情况（防御性）
					var fallbackMerchant models.Merchant
					if err := global.Db.Where("base_id = ?", chat.MerchantID).First(&fallbackMerchant).Error; err == nil {
						targetBaseID = fallbackMerchant.BaseID
					}
				}
			}
		} else {
			// 发送者是商家，目标为用户的 base_id
			targetBaseID = chat.UserBaseID
		}
		log.Println("🎯 targetBaseID =", targetBaseID)

		if targetBaseID != 0 {
			connMu.RLock()
			targetConn, ok := connStore[targetBaseID]
			connMu.RUnlock()
			log.Println("🔍 Find targetConn:", ok, "targetBaseID =", targetBaseID)
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
				log.Println("📤 Sending to", targetBaseID, "content =", chat.Content)
				if err := targetConn.WriteJSON(out); err != nil {
					log.Println("❌ WS WriteJSON failed:", err)
				} else {
					log.Println("✔ WS message delivered to", targetBaseID)
				}

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

// DebugConnections 返回当前活跃的 base_user id 列表，便于调试
func DebugConnections(c *gin.Context) {
	connMu.RLock()
	ids := make([]uint, 0, len(connStore))
	for k := range connStore {
		ids = append(ids, k)
	}
	connMu.RUnlock()
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": ids})
}
