package models

type RiderWallet struct {
	ID           uint    `gorm:"primaryKey" json:"id"`
	RiderID      uint    `json:"riderId"`
	Balance      float64 `json:"balance"`
	FrozenAmount float64 `json:"frozenAmount"`
	TotalIncome  float64 `json:"totalIncome"`
}
