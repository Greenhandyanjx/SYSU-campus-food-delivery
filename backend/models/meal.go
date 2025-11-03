package models

type Meal struct {
	ID          int     `gorm:"primaryKey"`
	Mealname    string  `gorm:"unique" json:"mealname"`
	Price       float64 `gorm:"not null" json:"price"`
	Description string  `gorm:"not null" json:"description"`
	MerchantID  int     `gorm:"not null" json:"merchant_id"`
	Status      string  `gorm:"not null" json:"status"`
	ImagePath   string  `gorm:"type:varchar(255)" json:"image_path"` // 存储图片相对路径
	Category    string  `gorm:"not null" json:"category"`
}