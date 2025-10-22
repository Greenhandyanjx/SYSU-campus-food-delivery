package models

type Dish struct {
	ID          int     `gorm:"primaryKey"`
	DishName    string  `gorm:"unique" json:"dish_name" form:"dish_name"`
	Price       float64 `gorm:"not null" json:"price" form:"price"`
	Description string  `gorm:"not null" json:"description" form:"description"`
	MerchantID  int     `gorm:"not null" json:"merchant_id" form:"merchant_id"`
	Tastes      string  `gorm:"not null" json:"tastes" form:"tastes"`
	ImagePath   string  `gorm:"type:varchar(255)" json:"image_path" form:"image_path"` // 存储图片相对路径
	Category    string  `gorm:"not null" json:"category" form:"category"`
}