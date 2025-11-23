package models

import "time"

type DeliveryRoute struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OrderID   uint      `json:"orderId"`
	RouteJSON string    `json:"route"` // [{"lat":..,"lng":..}]
	Distance  float64   `json:"distance"`
	ETA       int       `json:"estimatedTime"` // 预计时间（分钟）
	CreatedAt time.Time `json:"createdAt"`
}
