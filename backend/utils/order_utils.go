package utils

import (
	"backend/global"
	"backend/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// 转化函数
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

// 构建查询条件并实现分页查询
func FetchOrders(c *gin.Context, page, size int, beginTime, endTime time.Time, phonestr, numberstr, status string) ([]models.Order, int64, error) {
	// 构建查询条件
	query := global.Db.Model(&models.Order{})
	// 如果请求来自已认证的商家用户（中间件设置了 baseUserID），则根据 merchant.base_id 解析对应 merchant.id 并加入 merchant_id 过滤
	if baseIf, ok := c.Get("baseUserID"); ok {
		var baseID uint
		switch v := baseIf.(type) {
		case uint:
			baseID = v
		case int:
			baseID = uint(v)
		case int64:
			baseID = uint(v)
		case float64:
			baseID = uint(v)
		}
		if baseID != 0 {
			var m models.Merchant
			if err := global.Db.Where("base_id = ?", baseID).First(&m).Error; err == nil {
				query = query.Where("merchant_id = ?", m.ID)
			}
		}
	}
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
		// If requesting pending orders (status == 0) from merchant or rider endpoints, hide them.
		// This prevents merchants/riders from seeing user-initialized pending-pay orders.
		path := c.Request.URL.Path
		if stat == 0 && (strings.Contains(path, "/merchant/") || strings.Contains(path, "/rider/")) {
			// return empty result set
			return []models.Order{}, 0, nil
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
	// 查询总订单数 -- 保持和列表查询一致的过滤条件
	countQuery := global.Db.Model(&models.Order{})
	// 复制与上面相同的过滤器
	if baseIf, ok := c.Get("baseUserID"); ok {
		var baseID uint
		switch v := baseIf.(type) {
		case uint:
			baseID = v
		case int:
			baseID = uint(v)
		case int64:
			baseID = uint(v)
		case float64:
			baseID = uint(v)
		}
		if baseID != 0 {
			var m models.Merchant
			if err := global.Db.Where("base_id = ?", baseID).First(&m).Error; err == nil {
				countQuery = countQuery.Where("merchant_id = ?", m.ID)
			}
		}
	}
	if !beginTime.IsZero() {
		countQuery = countQuery.Where("created_at >= ?", beginTime)
	}
	if !endTime.IsZero() {
		countQuery = countQuery.Where("created_at <= ?", endTime)
	}
	if numberstr != "" {
		if num, err := strconv.Atoi(numberstr); err == nil {
			countQuery = countQuery.Where("ID = ?", num)
		}
	}
	if phonestr != "" {
		countQuery = countQuery.Where("phone = ?", phonestr)
	}
	if status != "" {
		if stat, err := strconv.Atoi(status); err == nil {
			// merchant/rider pending handling already done above for query; here mirror the same exclusion
			path := c.Request.URL.Path
			if stat == 0 && (strings.Contains(path, "/merchant/") || strings.Contains(path, "/rider/")) {
				// ensure count is zero as well
				count = 0
				return orders, count, nil
			}
			countQuery = countQuery.Where("status = ?", stat)
		}
	}
	countQuery.Count(&count)
	return orders, count, nil
}

// 查询consignee和address信息
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

// 赋值到新的结构体
func CopyOrdersToOrderWithDishnames(orders []models.Order, consigneeMap map[uint]models.Consignee, addressMap map[int]models.Address) []models.OrderWithDishnames {
	var ordersWithDetails []models.OrderWithDishnames
	for _, srcOrder := range orders {
		var dstOrder models.OrderWithDishnames
		dstOrder.ID = srcOrder.ID
		// ensure merchant id is preserved in the response for frontend secondary filtering
		dstOrder.MerchantID = srcOrder.MerchantID
		dstOrder.Ordertime = srcOrder.PickupPoint
		dstOrder.Dropofpoint = srcOrder.DropofPoint
		dstOrder.ExpectedTime = srcOrder.ExpectedTime
		// map finish time so frontend can display delivery/finish timestamp
		dstOrder.FinishAt = srcOrder.FinishAt
		// keep an alternative snake_case field mapping via JSON tag on the struct
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

// 获取菜品名称列表
func FetchDishnames(c *gin.Context, ordersWithDetails *[]models.OrderWithDishnames) {
	// 提取订单ID列表
	var orderIDs []uint
	for _, order := range *ordersWithDetails {
		orderIDs = append(orderIDs, order.ID)
	}
	// 获取 OrderDish 列表，并预加载 Dish
	var orderDishes []models.OrderDish
	result := global.Db.Preload("Dish").Where("order_id IN ?", orderIDs).Find(&orderDishes)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询订单菜品关联表失败", "data": nil})
		return
	}
	// 获取 OrderMeal 列表，并预加载 Meal
	var orderMeals []models.OrderMeal
	result1 := global.Db.Preload("Meal").Where("order_id IN ?", orderIDs).Find(&orderMeals)
	if result1.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询订单套餐关联表失败", "data": nil})
		return
	}
	// 构建 orderID 到 dishnames 的映射
	dishnamesMap := make(map[uint]string)
	for _, orderDish := range orderDishes {
		dishnamesMap[uint(orderDish.OrderID)] += fmt.Sprintf("%s*%d, ", orderDish.Dish.DishName, orderDish.Num)
	}
	for _, orderMeal := range orderMeals {
		dishnamesMap[uint(orderMeal.OrderID)] += fmt.Sprintf("%s*%d, ", orderMeal.Meal.Mealname, orderMeal.Num)
	}
	// 去除末尾的逗号和空格
	for orderID, dishnames := range dishnamesMap {
		dishnamesMap[orderID] = strings.TrimSuffix(dishnames, ", ")
	}
	// 将映射中的数据赋值给结构体表
	for i, order := range *ordersWithDetails {
		if dishnames, exists := dishnamesMap[order.ID]; exists {
			(*ordersWithDetails)[i].Orderdishes = dishnames
		} else {
			(*ordersWithDetails)[i].Orderdishes = ""
		}
	}
}
