package models

type CartItem struct {
	ID        uint    `gorm:"primaryKey"`
	CartID    uint    `gorm:"column:cart_id"`
	StoreID   string  `gorm:"column:store_id"`
	StoreName string  `gorm:"column:store_name"`
	DishID    string  `gorm:"column:dish_id"`
	Name      string  `gorm:"column:name"`
	Price     float64 `gorm:"column:price"`
	Qty       int64   `gorm:"column:qty"`
	Selected  bool    `gorm:"column:selected"`
	Category  string  `gorm:"column:category"`
}

func (CartItem) TableName() string {
	return "cart_items"
}
