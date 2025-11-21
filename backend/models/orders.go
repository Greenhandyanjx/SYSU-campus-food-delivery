package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model

	Consigneeid  int       `json:"consigneeid" gorm:"not null"`
	Phone        string    `json:"phone" gorm:"type:VARCHAR(20);not null"`
	Consignee    string    `json:"consignee" gorm:"type:VARCHAR(100);not null"`
	Address      string    `json:"address" gorm:"type:VARCHAR(255);not null"` // ← 你缺了这条！
	PickupPoint  time.Time `json:"pickuppoint" gorm:"type:datetime;not null;default:current_timestamp"`
	DropofPoint  time.Time `json:"dropofpoint" gorm:"type:datetime;not null;default:current_timestamp"`
	ExpectedTime time.Time `json:"expectedtime" gorm:"type:datetime;not null;default:current_timestamp"`
	Status       int       `json:"status" gorm:"not null default:'1'"`
	TotalPrice   float64   `json:"totalprice" gorm:"not null"`
	MerchantID   uint      `json:"merchantid" gorm:"not null"`
	Notes        string    `json:"notes"`

	Numberoftableware int `json:"numberoftableware" gorm:"not null default:'0'"`

	// ==== 新增骑手相关字段 ====
	RiderID    uint   `json:"rider_id" gorm:"default:0"`
	PickupCode string `json:"pickup_code"`
}
