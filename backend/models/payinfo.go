package models

import (
	"time"

	"gorm.io/gorm"
)

type PayInfo struct {
	gorm.Model
	Paymethod      int       `json:"paymethod" gorm:"not null"`
	Packamount     float64   `json:"packamount" gorm:"not null"`
	CheckoutTime   time.Time `json:"checkouttime" gorm:"not null"`
	Deliveryamount float64   `json:"deliveryamount" gorm:"not null;default:2"`

	// 支付对接相关字段
	OutTradeNo string     `json:"out_trade_no" gorm:"type:varchar(128);index"`
	CodeURL    string     `json:"code_url" gorm:"type:text"`
	Status     string     `json:"status" gorm:"type:varchar(32);default:'pending'"`
	PaidAt     *time.Time `json:"paid_at"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
