package models

import "time"

type RiderIncomeRecord struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RiderID   uint      `json:"riderId"`
	OrderID   uint      `json:"orderId"`
	Amount    float64   `json:"amount"` // 收入金额
	Type      string    `json:"type"`   // order | bonus | adjustment
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"createdAt"`
}
