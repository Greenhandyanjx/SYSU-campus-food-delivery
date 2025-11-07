package models

type User struct {
	BaseID   uint   `gorm:"uniqueIndex:idx_base_id"`
	Username string `gorm:"type:VARCHAR(255);not null;uniqueIndex"`
	Password string `gorm:"not null" json:"password"`
	Phone    string `gorm:"not null" json:"phone"`
}
