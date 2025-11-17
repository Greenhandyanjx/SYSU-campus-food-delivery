package models

import "gorm.io/gorm"

type MealDish struct {
	gorm.Model
	MealID int `gorm:"column:meal_id;not null"`
	DishID int `gorm:"column:dish_id;not null"`
	Num    int `gorm:"column:num;not null"`
}