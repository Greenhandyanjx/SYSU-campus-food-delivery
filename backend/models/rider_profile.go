package models

import "gorm.io/gorm"

type RiderProfile struct {
	gorm.Model

	UserID  uint `json:"userId"`  // 登录用户的 ID
	RiderID uint `json:"riderId"` // 关联实名认证 Rider 表

	Name            string  `json:"name"`
	Avatar          string  `json:"avatar"`
	Phone           string  `json:"phone"`
	Rating          float64 `json:"rating"`
	CompletedOrders int     `json:"completedOrders"`
	IsOnline        bool    `json:"isOnline"`

	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Address   string  `json:"address"`
}
