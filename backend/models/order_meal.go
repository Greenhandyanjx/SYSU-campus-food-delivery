package models

import (
	"gorm.io/gorm"
)

type OrderMeal struct {
	gorm.Model
	OrderID int `gorm:"not null"`
	MealID  int `gorm:"not null"`
	Num int `gorm:"not null"`
	Meal Meal `gorm:"foreignKey:MealID"` // 添加关联字段
} 