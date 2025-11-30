package models

type MerchantDeliveryConfig struct {
	BaseID        uint    `gorm:"primaryKey;column:base_id" json:"base_id"`
	MinPrice      float64 `gorm:"type:decimal(8,2);default:15" json:"min_price"`
	DeliveryFee   float64 `gorm:"type:decimal(8,2);default:2" json:"delivery_fee"`
	DeliveryRange int     `gorm:"default:2000" json:"delivery_range"`
}
