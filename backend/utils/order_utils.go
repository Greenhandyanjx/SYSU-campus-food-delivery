package utils

import (
	"backend/global"
	"backend/models"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

//转化函数
func ParsePaginationAndTime(c *gin.Context, pageStr, sizeStr, beginStr, endStr string) (int, int, time.Time, time.Time) {
    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        page = 1
    }
    size, err := strconv.Atoi(sizeStr)
    if err != nil || size < 1 {
        size = 20
    }
    var beginTime, endTime time.Time
    if beginStr != "" {
        beginTime, err = time.Parse("2006-01-02 15:04:05", beginStr)
        if err != nil {
            fmt.Println(err)
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid begin time format", "data": nil})
            return 0, 0, time.Time{}, time.Time{}
        }
    }
    if endStr != "" {
        endTime, err = time.Parse("2006-01-02 15:04:05", endStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid end time format", "data": nil})
            return 0, 0, time.Time{}, time.Time{}
        }
    }
    return page, size, beginTime, endTime
}

//构建查询条件并实现分页查询
func FetchOrders(c *gin.Context, page, size int, beginTime, endTime time.Time, phonestr, numberstr, status string) ([]models.Order, int64, error) {
    // 构建查询条件
    query := global.Db.Model(&models.Order{})
    if !beginTime.IsZero() {
        query = query.Where("created_at >= ?", beginTime)
    }
    if !endTime.IsZero() {
        query = query.Where("created_at <= ?", endTime)
    }
    if numberstr != "" {
        num, err := strconv.Atoi(numberstr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid number format", "data": nil})
            return nil, 0, err
        }
        query = query.Where("ID= ?", num)
    }
    if phonestr != "" {
        query = query.Where("phone = ?", phonestr)
    }
    if status != "" {
        stat, err := strconv.Atoi(status)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "invalid status format", "data": nil})
            return nil, 0, err
        }
        query = query.Where("status = ?", stat)
    }
    var orders []models.Order
    var count int64
    // 计算分页偏移量
    offset := (page - 1) * size
    // 查询订单列表
    result := query.Limit(size).Offset(offset).Find(&orders)
    if result.Error != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get order page", "data": nil})
        return nil, 0, result.Error
    }
    // 查询总订单数
    countQuery := global.Db.Model(&models.Order{})
    if !beginTime.IsZero() {
        countQuery = countQuery.Where("created_at >= ?", beginTime)
    }
    if !endTime.IsZero() {
        countQuery = countQuery.Where("created_at <= ?", endTime)
    }
    countQuery.Count(&count)
    return orders, count, nil
}

//查询consignee和address信息
func FetchConsigneesAndAddresses(c *gin.Context, orders []models.Order) (map[uint]models.Consignee, map[int]models.Address) {
    // 提取 ConsigneeID 列表
    var consigneeIDs []uint
    for _, order := range orders {
        consigneeIDs = append(consigneeIDs, uint(order.Consigneeid))
    }
    // 查询 Consignee 列表
    var consignees []models.Consignee
    if err := global.Db.Where("id IN ?", consigneeIDs).Find(&consignees).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get consignees", "data": nil})
        return nil, nil
    }
    // 构建 ConsigneeID 和 Consignee 的映射
    consigneeMap := make(map[uint]models.Consignee)
    for _, consignee := range consignees {
        consigneeMap[consignee.ID] = consignee
    }
    // 提取 AddressID 列表
    var addressIDs []int
    for _, consignee := range consignees {
        addressIDs = append(addressIDs, consignee.Addressid)
    }
    // 查询 Address 列表
    var addresses []models.Address
    if err := global.Db.Where("id IN ?", addressIDs).Find(&addresses).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "failed to get addresses", "data": nil})
        return nil, nil
    }
    // 构建 AddressID 和 Address 的映射
    addressMap := make(map[int]models.Address)
    for _, address := range addresses {
        addressMap[address.ID] = address
    }
    return consigneeMap, addressMap
}

//赋值到新的结构体
func CopyOrdersToOrderWithDishnames(orders []models.Order, consigneeMap map[uint]models.Consignee, addressMap map[int]models.Address) []models.OrderWithDishnames {
    var ordersWithDetails []models.OrderWithDishnames
    for _, srcOrder := range orders {
        var dstOrder models.OrderWithDishnames
        dstOrder.ID = srcOrder.ID
        dstOrder.Ordertime=srcOrder.PickupPoint
        dstOrder.Dropofpoint = srcOrder.DropofPoint
        dstOrder.ExpectedTime = srcOrder.ExpectedTime
        dstOrder.Status = srcOrder.Status
        dstOrder.TablewareNumber = srcOrder.Numberoftableware
        dstOrder.TotalPrice = srcOrder.TotalPrice
        dstOrder.Remark = srcOrder.Notes
        dstOrder.ConsigneeID = uint(srcOrder.Consigneeid)
        // 获取 Consignee 信息
        consignee, consigneeExists := consigneeMap[dstOrder.ConsigneeID]
        if consigneeExists {
            dstOrder.ConsigneeName = consignee.Name
            dstOrder.Phone = consignee.Phone
            dstOrder.ConsigneeAddressID = consignee.Addressid
            // 获取 Address 信息并拼接完整的地址
            address, addressExists := addressMap[consignee.Addressid]
            if addressExists {
                dstOrder.Address = fmt.Sprintf("%s %s %s %s %s", address.Province, address.City, address.District, address.Street, address.Detail)
            }
        }
        ordersWithDetails = append(ordersWithDetails, dstOrder)
    }
    return ordersWithDetails
}

//获取菜品名称列表
func FetchDishnames(c *gin.Context, ordersWithDetails *[]models.OrderWithDishnames) {
    // 提取订单ID列表
    var orderIDs []uint
    for _, order := range *ordersWithDetails {
        orderIDs = append(orderIDs, order.ID)
    }
    // 查询 orderdish 表以获取每个订单的菜品名
    var orderDishnames []struct {
        OrderID   uint     `gorm:"column:order_id"`
        Dishnames string   `gorm:"column:dishnames"` // 修改为 string 类型，因为 GROUP_CONCAT 返回的是一个字符串
    }
    if err := global.Db.Table("order-dish").
        Select("order_id, GROUP_CONCAT(dishname) as dishnames").
        Where("order_id IN ?", orderIDs).
        Group("order_id").
        Find(&orderDishnames).Error; err != nil {
        log.Printf("查询订单菜品名失败: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询订单菜品名失败", "data": nil})
        return
    }
    // 构建订单和菜品名的映射
    dishnamesMap := make(map[uint]string)
    for _, od := range orderDishnames {
        dishnamesMap[od.OrderID] = od.Dishnames
    }
    // 将映射中的数据赋值给结构体表
    for i, order := range *ordersWithDetails {
        if dishnames, exists := dishnamesMap[order.ID]; exists {
            (*ordersWithDetails)[i].Orderdishes = dishnames
        }
    }
}