// backend/models/rider_extra.go
package models

import "time"

// 异常上报
type RiderIssue struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	OrderID     uint       `json:"orderId"`
	RiderID     uint       `json:"riderId"`
	Type        string     `json:"type"`        // 如：traffic、customer、restaurant
	Description string     `json:"description"` // 异常描述
	Images      string     `json:"images"`      // JSON 数组字符串：["url1","url2"]
	Timestamp   *time.Time `json:"timestamp"`   // 前端传的时间（可空）
	CreatedAt   time.Time  `json:"createdAt"`
}

// 骑手评价
type RiderReview struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RiderID   uint      `json:"riderId"`
	OrderID   uint      `json:"orderId"`
	Rating    float64   `json:"rating"`
	Comment   string    `json:"comment"`
	CreatedAt time.Time `json:"createdAt"`
}

// 通知
type RiderNotification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	RiderID   uint      `json:"riderId"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Type      string    `json:"type"`   // order | system | marketing ...
	IsRead    bool      `json:"isRead"` // 是否已读
	CreatedAt time.Time `json:"createdAt"`
}

// 系统消息（全局）
type RiderSystemMessage struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Type        string    `json:"type"` // 通告类型
	PublishedAt time.Time `json:"publishedAt"`
}

// 工作设置
type RiderWorkSettings struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	RiderID       uint   `json:"riderId"`
	AutoAccept    bool   `json:"autoAccept"`
	DeliveryRange int    `json:"deliveryRange"` // 配送范围（km）
	WorkTimeStart string `json:"-"`             // "09:00"
	WorkTimeEnd   string `json:"-"`
	RestEnabled   bool   `json:"-"`
	RestStart     string `json:"-"`
	RestEnd       string `json:"-"`
	MaxOrders     int    `json:"maxOrders"`
}

// 账户设置
type RiderAccountSettings struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	RiderID  uint   `json:"riderId"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Wechat   string `json:"wechat"`
	Alipay   string `json:"alipay"`
	BankCard string `json:"bankCard"`
}

// 通知设置
type RiderNotificationSettings struct {
	ID                 uint `gorm:"primaryKey" json:"id"`
	RiderID            uint `json:"riderId"`
	OrderNotification  bool `json:"orderNotification"`
	SystemNotification bool `json:"systemNotification"`
	SoundEnabled       bool `json:"soundEnabled"`
	VibrationEnabled   bool `json:"vibrationEnabled"`
}

// 实名认证
type RiderVerification struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	RiderID     uint       `json:"riderId"` // 这里用 base_user_id 做 RiderID
	RealName    string     `json:"realName"`
	IdCard      string     `json:"idCard"`
	IdCardFront string     `json:"idCardFront"`
	IdCardBack  string     `json:"idCardBack"`
	HealthCert  string     `json:"healthCert"`
	Status      string     `json:"status"`     // pending | approved | rejected
	SubmitTime  time.Time  `json:"submitTime"` // 提交时间
	ReviewTime  *time.Time `json:"reviewTime"`
}

// 热力图数据点
type RiderHeatmapPoint struct {
	ID      uint      `gorm:"primaryKey" json:"id"`
	RiderID uint      `json:"riderId"`
	Lat     float64   `json:"lat"`
	Lng     float64   `json:"lng"`
	Count   int       `json:"count"`
	Date    time.Time `json:"date"` // 按天汇总
}
