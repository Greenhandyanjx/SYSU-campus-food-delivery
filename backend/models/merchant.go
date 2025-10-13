package models

type Merchant struct {
	ID        int    `json:"id"`
	Username  string `json:"name"`
	Status    string `json:"status"`
	MenuCount int    `json:"menu_count"`
	Password  string `json:"password"`
}