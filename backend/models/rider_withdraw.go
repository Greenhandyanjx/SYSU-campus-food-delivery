package models

import "time"

type RiderWithdraw struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	RiderID     uint       `json:"riderId"`
	Amount      float64    `json:"amount"`
	Account     string     `json:"account"`
	Status      string     `json:"status"` // pending | success | failed
	AppliedAt   time.Time  `json:"appliedAt"`
	ProcessedAt *time.Time `json:"processedAt"` // 允许为空
}
