package models

type User struct {
	BaseID   uint   `gorm:"unique"`
	Username string `gorm:"not null;unique" json:"username"`
	Password string `gorm:"not null" json:"password"`
	Phone    string `gorm:"not null;unique" json:"phone"`
}
