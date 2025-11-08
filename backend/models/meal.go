package models

import "gorm.io/gorm"

type Meal struct {
	ID          int    `gorm:"primaryKey"`
	Mealname    string `gorm:"unique" json:"name"`
	Price       string `gorm:"not null" json:"price"`
	Description string `gorm:"not null" json:"description"`
	MerchantID  uint    `gorm:"not null" json:"merchant_id"`
	Status      int    `gorm:"not null;default:1"  json:"status"`
	ImagePath   string `gorm:"type:varchar(255)" json:"image"` // 存储图片相对路径
	Category    int `gorm:"not null" json:"categoryId"`
	gorm.Model
}