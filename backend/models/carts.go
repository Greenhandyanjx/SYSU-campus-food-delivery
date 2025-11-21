package models

import (
	"time"

	"gorm.io/gorm"
)

type Cart struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    string `gorm:"column:user_id"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Cart) TableName() string {
	return "cart"
}
