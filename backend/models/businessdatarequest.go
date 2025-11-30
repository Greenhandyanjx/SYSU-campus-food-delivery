package models

type BusinessDataRequest struct {
	BaseID string `json:"baseid"` // 定义 baseid 字段
	Date   string `json:"date"`
}