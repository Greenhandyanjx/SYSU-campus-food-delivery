package models

import "gorm.io/gorm"

type OrderDish struct {
	gorm.Model
	OrderID int `gorm:"not null"`
	DishID  int `gorm:"not null"`
	Num int `gorm:"not null"`
}