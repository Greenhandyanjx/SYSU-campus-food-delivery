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
}
