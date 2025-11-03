package models

type Merchant struct {
	BaseID    uint   `gorm:"primary_key"`
	Username  string `json:"username"`
	Status    string `json:"status"`
	MenuCount int    `json:"menu_count"`
	Password  string `json:"password"`
}