package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderID uint `json:"orderid" gorm:"not null;unique"`
	PickupPoint string `json:"pickuppoint" gorm:"not null"`
	DropofPoint string `json:"dropofpoint" gorm:"not null"`
	ExpectedTime string `json:"expectedtime" gorm:"not null"`
	Status string `json:"status" gorm:"not null default:'pending'"`
	TotalPrice float64 `json:"totalprice" gorm:"not null"`
	MerchantID uint `json:"merchantid" gorm:"not null"`
}