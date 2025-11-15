package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	
	Consignee string `json:"consignee" gorm:"not null"`
	Phone string `json:"phone" gorm:"not null"`
	Address string `json:"address" gorm:"not null"`
	PickupPoint time.Time `json:"pickuppoint" gorm:"type:datetime;not null;default:current_timestamp"`
	DropofPoint time.Time `json:"dropofpoint" gorm:"type:datetime;not null;default:current_timestamp"`
	ExpectedTime time.Time `json:"expectedtime" gorm:"type:datetime;not null;default:current_timestamp"`
	Status int `json:"status" gorm:"not null default:'1'"`
	TotalPrice float64 `json:"totalprice" gorm:"not null"`
	MerchantID uint `json:"merchantid" gorm:"not null"`
}