package models

import (
	"time"
)

//定义新的结构体，包含原有的 order 属性和 dishnames 字段
type OrderWithDishnames struct {
    ID                  uint      `json:"orderid"`
    Ordertime time.Time `json:"orderTime"`
    Dropofpoint         time.Time `json:"dropofpoint"`
    ExpectedTime        time.Time `json:"expectedtime"`
    Status              int       `json:"status"`
    TablewareNumber   int       `json:"tablewareNumber"`
    TotalPrice          float64   `json:"totalprice"`
    Orderdishes           string    `json:"orderDishes"`
    Remark               string    `json:"remark"`
    ConsigneeID         uint      `json:"consigneeid"`
    ConsigneeName       string    `json:"consignee"`
    Phone    string    `json:"phone"`
    ConsigneeAddressID  int       `json:"consigneeaddressid"`
    Address string `json:"address"`
}