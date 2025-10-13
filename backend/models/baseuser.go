package models

import "gorm.io/gorm"

type BaseUser struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string `gorm:"not null"`
	Role     string `gorm:"not null"`
}