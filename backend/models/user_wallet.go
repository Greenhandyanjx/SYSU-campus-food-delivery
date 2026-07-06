package models

import (
	"time"

	"gorm.io/gorm"
)

// UserWallet 用户钱包（普通用户，非骑手）
type UserWallet struct {
	gorm.Model
	UserID  uint    `gorm:"uniqueIndex;not null" json:"user_id"`
	Balance float64 `gorm:"type:decimal(10,2);default:0" json:"balance"`
}

// WalletTransaction 钱包交易流水
type WalletTransaction struct {
	gorm.Model
	UserID        uint      `gorm:"index;not null" json:"user_id"`
	Amount        float64   `gorm:"type:decimal(10,2);not null" json:"amount"`
	Type          string    `gorm:"type:varchar(32);not null" json:"type"` // "recharge" / "payment" / "refund"
	BalanceBefore float64   `gorm:"type:decimal(10,2)" json:"balance_before"`
	BalanceAfter  float64   `gorm:"type:decimal(10,2)" json:"balance_after"`
	Description   string    `gorm:"type:varchar(255)" json:"description"`
	OrderID       int       `json:"order_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
}
