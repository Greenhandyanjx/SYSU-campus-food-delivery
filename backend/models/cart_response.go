package models

// 每家店的数据结构（shops 数组里的元素）
type CartShopResponse struct {
	StoreID string             `json:"storeId"`
	Name    string             `json:"name"`
	Items   []CartItemResponse `json:"items"`
}

// 单个 cartItem 的数据结构（Items数组里的元素）
type CartItemResponse struct {
	DishID      string  `json:"dishId"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Qty         int64   `json:"qty"`
	OriginalQty int64   `json:"originalQty"`
	Selected    bool    `json:"selected"`
	Category    string  `json:"category,omitempty"`
}
