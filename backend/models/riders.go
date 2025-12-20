package models

type Rider struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	BaseID   uint   `gorm:"uniqueIndex;not null" json:"base_id"`
	RealName string `gorm:"column:realname" json:"realname"`
	IDNumber string `gorm:"column:idnumber" json:"idnumber"`
	Phone    string `gorm:"type:VARCHAR(20);unique" json:"phone"`
	// 评分统计
	AvgScore   float64 `json:"avg_score" gorm:"type:decimal(3,2);default:4.0"`
	ScoreCount int     `json:"score_count" gorm:"default:1"`
}
