package models

import (
	"time"

	"gorm.io/gorm"
)

type PayInfo struct {
	gorm.Model
	Paymethod int `json:"paymethod" gorm:"not null"`
	Packamount float64 `json:"packamount" gorm:"not null"`
    CheckoutTime time.Time `json:"checkouttime" gorm:"not null"`
	Deliveryamount float64 `json:"deliveryamount" gorm:"not null"`
}