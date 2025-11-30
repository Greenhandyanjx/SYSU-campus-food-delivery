package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	BaseID   uint   `gorm:"uniqueIndex;not null" json:"base_id"`
	Nickname string `json:"nickname"`
	Phone    string `gorm:"type:VARCHAR(20);unique" json:"phone"`
	Address  string `gorm:"type:VARCHAR(512)" json:"address"`
}
