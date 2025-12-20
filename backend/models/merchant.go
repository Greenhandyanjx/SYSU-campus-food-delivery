package models

type Merchant struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	BaseID       uint   `gorm:"uniqueIndex;not null" json:"base_id"`
	ShopName     string `json:"shop_name"`
	ShopLocation string `json:"shop_location"`
	Owner        string `json:"owner"`
	Phone        string `gorm:"type:VARCHAR(20)" json:"phone"`
	Logo         string `json:"logoUrl"`
	License      string `json:"licenseUrl"`
	Status       string `json:"status"`
	MenuCount    int    `json:"menu_count"`
	// TopCategory1/2 存储商家最常见的两个菜品/套餐分类 ID（1-15）
	TopCategory1 int `json:"top_category_1" gorm:"default:0"`
	TopCategory2 int `json:"top_category_2" gorm:"default:0"`
	// 评分统计：平均分（1.0-5.0）与评分人数
	AvgScore   float64 `json:"avg_score" gorm:"type:decimal(3,2);default:4.0"`
	ScoreCount int     `json:"score_count" gorm:"default:1"`
}
