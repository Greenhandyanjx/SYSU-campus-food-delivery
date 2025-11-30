package models

import "gorm.io/gorm"

type BaseUser struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Role     string `gorm:"not null" json:"role" binding:"required"`
}
