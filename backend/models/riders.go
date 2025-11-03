package models

type Rider struct {
	BaseID   uint   `gorm:"primary_key"`
	Username string `gorm:"not null;unique" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Phone    string `gorm:"not null;unique" json:"phone"`
}