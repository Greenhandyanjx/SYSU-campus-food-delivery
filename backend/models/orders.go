package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderID uint `json:"orderid" gorm:"not null;unique"`
	PickupPoint time.Time `json:"pickuppoint" gorm:"not null"`
	DropofPoint time.Time `json:"dropofpoint" gorm:"not null"`
	ExpectedTime time.Time `json:"expectedtime" gorm:"not null"`
	Status int `json:"status" gorm:"not null default:'1'"`
	TotalPrice float64 `json:"totalprice" gorm:"not null"`
	MerchantID uint `json:"merchantid" gorm:"not null"`
}