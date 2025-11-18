package models

type Rider struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	BaseID   uint   `gorm:"uniqueIndex;not null" json:"base_id"`
	RealName string `gorm:"column:realname" json:"realname"`
	IDNumber string `gorm:"column:idnumber" json:"idnumber"`
	Phone    string `gorm:"type:VARCHAR(20);unique" json:"phone"`
}
