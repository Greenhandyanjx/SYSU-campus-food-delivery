package models

import (
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	OrderID uint `json:"orderid" gorm:"not null"`
	PickupPoint string `json:"pickuppoint" gorm:"not null"`
	DropofPoint string `json:"dropofpoint" gorm:"not null"`
	ExpectedTime string `json:"expectedtime" gorm:"not null"`
	Status string `json:"status" gorm:"not null"`
	TotalPrice float64 `json:"totalprice" gorm:"not null"`
}