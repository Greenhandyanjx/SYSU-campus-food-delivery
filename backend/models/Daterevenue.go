package models

import (
	"time"
)

type Revenue struct {
    ID        uint      `gorm:"primary_key"` // 主键
    MerchantID uint      `json:"merchant_id"` // 商家ID
    Revenue   float64   `json:"revenue"`     // 营业额
    Day      time.Time `json:"day"`        // 日期
    Usernumber int `gorm:"column:usernumber" json:"usernumber"`
}