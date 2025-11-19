package models

import "time"

// ChatMessage 表示用户与商家之间的一条聊天消息
type ChatMessage struct {
	ID uint `gorm:"primaryKey" json:"id"`
	// 发送者 base_user id
	FromBaseID uint `json:"from_base_id"`
	// 如果发送者是用户，这里存用户的 base_id；如果发送者是商家，也存商家对应的 base_id
	// 接收者的商家 id（merchants 表的主键）
	MerchantID uint `json:"merchant_id" index:"idx_merchant_user"`
	// 对话中的用户 base_id（便于查询）
	UserBaseID uint `json:"user_base_id" index:"idx_merchant_user"`

	// 消息内容（文本或图片 URL）
	Content string `json:"content" gorm:"type:text"`
	// 类型：text/image/other
	Type string `json:"type" gorm:"size:32"`
	// 状态：sent/delivered/read
	Status string `json:"status" gorm:"size:32"`

	CreatedAt   time.Time  `json:"created_at"`
	DeliveredAt *time.Time `json:"delivered_at"`
	ReadAt      *time.Time `json:"read_at"`
}
