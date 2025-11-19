package models

import "time"

// SalesStat 按商家/日期/类型汇总销量（dish 或 meal）
type SalesStat struct {
    ID         uint      `gorm:"primaryKey" json:"id"`
    MerchantID uint      `gorm:"not null;index:idx_sales,priority:1" json:"merchant_id"`
    ItemType   string    `gorm:"type:varchar(16);not null;index:idx_sales,priority:2" json:"item_type"` // "dish" or "meal"
    ItemID     uint      `gorm:"not null;index:idx_sales,priority:3" json:"item_id"`
	Itemname string `gorm:"type:varchar(128);not null" json:"itemname"`
    Date       time.Time `gorm:"type:date;not null;index:idx_sales,priority:4" json:"date"`
    Quantity   int    `gorm:"not null;default:0" json:"quantity"`
    Revenue    float64   `gorm:"not null;default:0" json:"revenue"`
    CreatedAt  time.Time `json:"created_at"`
    UpdatedAt  time.Time `json:"updated_at"`
}

// // OrderSalesProcessed 用于记录哪些订单的销量已经计入（幂等）
// type OrderSalesProcessed struct {
//     ID        uint      `gorm:"primaryKey" json:"id"`
//     OrderID   uint      `gorm:"uniqueIndex;not null" json:"order_id"`
//     CreatedAt time.Time `json:"created_at"`
// }