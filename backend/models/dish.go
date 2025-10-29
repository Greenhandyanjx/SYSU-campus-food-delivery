package models

import "gorm.io/gorm"

type Dish struct {
	
	ID          int      `gorm:"primaryKey"`
	DishName    string   `gorm:"unique" json:"name" form:"dish_name"`
	Price       string   `gorm:"not null" json:"price" form:"price"`
	Description string   `gorm:"not null" json:"description" form:"description"`
	MerchantID  int      `gorm:"not null" json:"merchant_id" form:"merchant_id"`
	Tastes      string   `gorm:"not null" json:"tastes" form:"tastes"`
	ImagePath   string   `gorm:"type:varchar(255)" json:"image" form:"image_path"` // 存储图片相对路径
	Category    int      `gorm:"not null" json:"categoryId" form:"category"`
	Flavors     []Flavor `json:"flavors"`
	Status      int      `gorm:"not null;default:1" json:"status" form:"status"`
	gorm.Model
}

func (Dish) TableName() string {
    return "dishes"
}


type Flavor struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	Name   string `json:"name" gorm:"column:flavor_name"`
	Value  string `json:"value" gorm:"column:flavor_value"`
	DishID int    `gorm:"foreignKey:DishID" json:"dishId"` // 外键字段
}

func (Flavor) TableName() string {
    return "flavors"
}