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
	PickupPoint  time.Time `json:"pickuppoint" gorm:"type:datetime;not null;default:current_timestamp"`  //接单时间
	DropofPoint  time.Time `json:"dropofpoint" gorm:"type:datetime;not null;default:current_timestamp"`  //送达时间
	ExpectedTime time.Time `json:"expectedtime" gorm:"type:datetime;not null;default:current_timestamp"` //期望时间
	Status       int       `json:"status" gorm:"not null default:'1'"`
	TotalPrice   float64   `json:"totalprice" gorm:"not null"`
	MerchantID   uint      `json:"merchantid" gorm:"not null"`
	Notes        string    `json:"notes"`
	//餐具数量
	Numberoftableware int `json:"numberoftableware" gorm:"not null default:'0'"`

	// ==== 新增骑手相关字段 ====
	RiderID    uint   `json:"rider_id" gorm:"default:0"`
	PickupCode string `json:"pickup_code"`
}
