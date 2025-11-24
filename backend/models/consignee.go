package models

type Consignee struct {
	ID        uint   `gorm:"primaryKey" json:"id"`
	Userid    uint   `gorm:"not null" json:"userid"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Addressid int    `json:"addressid"`
}