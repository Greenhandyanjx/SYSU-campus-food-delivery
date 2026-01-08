package models

type CartItem struct {
	ID         uint   `json:"id"`
	CartID     uint   `json:"cart_id"`
	MerchantID uint   `json:"merchant_id"`
	DishID     uint   `json:"dish_id"`
	MealID     uint   `json:"meal_id"` // 支持套餐（meal）加入购物车
	Name       string `json:"name"`
	Price      string `gorm:"not null;type:varchar(20)" json:"price"`
	Qty        int    `json:"qty"`
	Selected   bool   `json:"selected"`
}

func (CartItem) TableName() string {
	return "cart_items"
}
