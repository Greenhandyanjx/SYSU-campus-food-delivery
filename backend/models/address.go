package models

type Address struct {
	ID       int    `json:"id"`
	UserID   int    `json:"user_id"`
	Province string `json:"province"`
	City     string `json:"city"`
	District string `json:"district"`
	Street   string `json:"street"`
	Detail   string `json:"detail"`
}