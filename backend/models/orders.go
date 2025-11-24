package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	gorm.Model

<<<<<<< HEAD
	Consigneeid int `json:"consigneeid" gorm:"not null"`

=======
	Consigneeid  int       `json:"consigneeid" gorm:"not null"`
	
>>>>>>> dev
	PickupPoint  time.Time `json:"pickuppoint" gorm:"type:datetime;not null;default:current_timestamp"`
	DropofPoint  time.Time `json:"dropofpoint" gorm:"type:datetime;not null;default:current_timestamp"`
	ExpectedTime time.Time `json:"expectedtime" gorm:"type:datetime;not null;default:current_timestamp"`
	Status       int       `json:"status" gorm:"not null default:'1'"`
	TotalPrice   float64   `json:"totalprice" gorm:"not null"`
	MerchantID   uint      `json:"merchantid" gorm:"not null"`
	Notes        string    `json:"notes"`

	Numberoftableware int `json:"numberoftableware" gorm:"not null default:'0'"`

	//支付信息
	PayInfoid int     `json:"payid" gorm:"not null"`
	PayInfo   PayInfo `gorm:"foreignKey:PayInfoid"`
	// ==== 新增骑手相关字段 ====
	RiderID    uint   `json:"rider_id" gorm:"default:0"`
	PickupCode string `json:"pickup_code"`

	// ========== 时间线字段（前端强需求） ==========
	AcceptedAt *time.Time `json:"acceptedAt"` // 接单时间
	PickupAt   *time.Time `json:"pickupAt"`   // 取货时间
	DeliverAt  *time.Time `json:"deliverAt"`  // 开始配送时间
	FinishAt   *time.Time `json:"finishAt"`   // 完成时间 （已派送）

}
