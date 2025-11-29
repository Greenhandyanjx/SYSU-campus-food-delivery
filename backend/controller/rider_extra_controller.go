package controller

import (
	"backend/global"
	"backend/models"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ========== 配送状态扩展接口 ==========

// PUT /rider/orders/:orderId/start
// startDelivery()
func StartDelivery(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	orderId := c.Param("orderId")

	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	// 允许从待取货(2) 或 配送中(3) 开始
	if order.Status != 2 && order.Status != 3 {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "订单状态不正确"})
		return
	}

	now := time.Now()
	if err := global.Db.Model(&order).Updates(map[string]interface{}{
		"status":     3,
		"pickup_at":  gorm.Expr("IF(pickup_at IS NULL, ?, pickup_at)", now),
		"updated_at": now,
	}).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "开始配送失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"success":   true,
			"startTime": now,
		},
	})
}

// PUT /rider/orders/:orderId/arrive-pickup
// arrivePickup()
func ArrivePickup(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	orderId := c.Param("orderId")

	var req struct {
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		Code      string   `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空 body
	}

	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	// 校验取餐码（如果前端传了）
	if req.Code != "" && order.PickupCode != "" && req.Code != order.PickupCode {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "取餐码不正确"})
		return
	}

	now := time.Now()
	// 可选更新骑手位置
	if req.Latitude != nil && req.Longitude != nil {
		global.Db.Model(&models.RiderProfile{}).
			Where("user_id = ?", riderID).
			Updates(map[string]interface{}{
				"latitude":  *req.Latitude,
				"longitude": *req.Longitude,
			})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{"success": true, "arrivedAt": now},
	})
}

// PUT /rider/orders/:orderId/status
// updateDeliveryStatus()
func UpdateDeliveryStatus(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	orderId := c.Param("orderId")

	var req struct {
		Status    string   `json:"status"` // delivering | completed | exception ...
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
		Note      string   `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	var order models.Order
	if err := global.Db.Where("id = ? AND rider_id = ?", orderId, riderID).First(&order).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "订单不存在或无权限"})
		return
	}

	updates := map[string]interface{}{}
	switch req.Status {
	case "delivering":
		updates["status"] = 3
	case "completed":
		updates["status"] = 4
		now := time.Now()
		updates["finish_at"] = &now
		updates["dropof_point"] = now
	}

	if len(updates) > 0 {
		if err := global.Db.Model(&order).Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "状态更新失败"})
			return
		}
	}

	// 更新骑手位置
	if req.Latitude != nil && req.Longitude != nil {
		global.Db.Model(&models.RiderProfile{}).
			Where("user_id = ?", riderID).
			Updates(map[string]interface{}{
				"latitude":  *req.Latitude,
				"longitude": *req.Longitude,
			})
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}

// POST /rider/orders/:orderId/issue
// reportIssue()
func ReportIssue(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	orderId := c.Param("orderId")

	var req struct {
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Images      []string `json:"images"`
		Timestamp   *int64   `json:"timestamp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	oid, _ := strconv.Atoi(orderId)
	var ts *time.Time
	if req.Timestamp != nil {
		t := time.Unix(*req.Timestamp, 0)
		ts = &t
	}

	imagesJSON, _ := json.Marshal(req.Images)

	issue := models.RiderIssue{
		OrderID:     uint(oid),
		RiderID:     riderID,
		Type:        req.Type,
		Description: req.Description,
		Images:      string(imagesJSON),
		Timestamp:   ts,
		CreatedAt:   time.Now(),
	}
	if err := global.Db.Create(&issue).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"success": true,
			"issueId": issue.ID,
		},
	})
}

// ========== 收入明细 & 配送记录 ==========

// GET /rider/income/details
// getIncomeDetails()
func GetIncomeDetails(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = 20
	}
	offset := (page - 1) * size

	itype := c.Query("type")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var records []models.RiderIncomeRecord
	var total int64

	query := global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ?", riderID)

	if itype != "" {
		query = query.Where("type = ?", itype)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	query.Count(&total)

	if err := query.
		Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&records).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	// 和 GetIncomeHistory 一样，手动映射字段名
	items := make([]gin.H, 0, len(records))
	for _, r := range records {
		items = append(items, gin.H{
			"id":      r.ID,
			"orderId": r.OrderID,
			"amount":  r.Amount,
			"type":    r.Type,
			"time":    r.CreatedAt, // ⭐ 关键：time 字段
			"remark":  r.Remark,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items": items,
			"total": total,
		},
	})
}

// GET /rider/delivery/records
// getDeliveryRecords()
func GetDeliveryRecords(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	offset := (page - 1) * size

	status := c.Query("status") // 暂时不用，可扩展
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	type Record struct {
		ID          uint      `json:"id"`
		OrderNo     uint      `json:"orderNo"`
		Distance    float64   `json:"distance"`
		Duration    int64     `json:"duration"`
		CompletedAt time.Time `json:"completedAt"`
	}

	var list []Record
	var total int64

	db := global.Db.Table("orders").
		Select(`orders.id AS id,
		        orders.id AS order_no,
		        COALESCE(delivery_routes.distance, 0) AS distance,
		        TIMESTAMPDIFF(MINUTE, orders.pickup_at, orders.finish_at) AS duration,
		        orders.finish_at AS completed_at`).
		Joins("LEFT JOIN delivery_routes ON delivery_routes.order_id = orders.id").
		Where("orders.rider_id = ? AND orders.status = 4", riderID)

	if status != "" {
		// 未来可根据 status 映射
	}
	if startDate != "" {
		db = db.Where("orders.finish_at >= ?", startDate)
	}
	if endDate != "" {
		db = db.Where("orders.finish_at <= ?", endDate)
	}

	db.Count(&total)
	db.Order("orders.finish_at DESC").
		Offset(offset).
		Limit(size).
		Scan(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items": list,
			"total": total,
		},
	})
}

// ========== 工作统计相关 ==========

// GET /rider/stats/work
// getWorkStats()
func GetWorkStats(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	period := c.DefaultQuery("period", "today")

	now := time.Now()
	var start time.Time
	switch period {
	case "today":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		start = now.AddDate(0, 0, -7)
	case "month":
		start = now.AddDate(0, -1, 0)
	default:
		start = now.AddDate(0, 0, -7)
	}

	var totalIncome float64
	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, start).
		Select("SUM(amount)").Scan(&totalIncome)

	var totalOrders int64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Count(&totalOrders)

	// 计算在线天数（有订单记录的天数）
	type Day struct {
		Day time.Time
	}
	var days []Day
	global.Db.Model(&models.Order{}).
		Select("DATE(finish_at) AS day").
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Group("DATE(finish_at)").Scan(&days)

	// 计算平均配送时间（分钟）
	var avgDeliveryTime float64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Select("AVG(TIMESTAMPDIFF(MINUTE, created_at, finish_at))").Scan(&avgDeliveryTime)

	// 计算平均配送距离（公里）
	var avgDistance float64
	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Select("AVG(distance)").Scan(&avgDistance)

	// 计算在线时长（小时） - 按有订单的时间估算
	var onlineHours float64 = float64(len(days)) * 8 // 假设每天工作8小时

	// 计算效率（单/小时）
	var efficiency float64
	if onlineHours > 0 {
		efficiency = float64(totalOrders) / onlineHours
	}

	// 暂时把完成率视为 100%，有更多数据源可以再优化
	completionRate := 100.0
	if totalOrders == 0 {
		completionRate = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"completedOrders":     totalOrders,
			"cancelledOrders":     0, // 暂时固定为0，可后续扩展
			"totalIncome":        totalIncome,
			"completionRate":     completionRate,
			"onlineHours":        onlineHours,
			"workDays":          len(days),
			"avgDeliveryTime":    int(avgDeliveryTime),
			"avgDistance":        avgDistance,
			"efficiency":        efficiency,
		},
	})
}

// GET /rider/stats/monthly
// getMonthlyStats()
func GetMonthlyStats(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var monthIncome float64
	var monthOrders int64

	global.Db.Model(&models.RiderIncomeRecord{}).
		Where("rider_id = ? AND created_at >= ?", riderID, start).
		Select("SUM(amount)").Scan(&monthIncome)

	global.Db.Model(&models.Order{}).
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Count(&monthOrders)

	// 统计在线天数：有完成订单的那几天
	type Day struct {
		Day time.Time
	}
	var days []Day
	global.Db.Model(&models.Order{}).
		Select("DATE(finish_at) AS day").
		Where("rider_id = ? AND status = 4 AND finish_at >= ?", riderID, start).
		Group("DATE(finish_at)").Scan(&days)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"monthOrders": monthOrders,
			"monthIncome": monthIncome,
			"onlineDays":  len(days),
		},
	})
}

// ========== 评价 & 排行榜 ==========

// GET /rider/reviews
// getReviews()
func GetReviews(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	offset := (page - 1) * size

	var list []models.RiderReview
	var total int64

	query := global.Db.Model(&models.RiderReview{}).
		Where("rider_id = ?", riderID)

	query.Count(&total)
	query.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&list)

	// 计算平均评分
	var avgRating float64
	global.Db.Model(&models.RiderReview{}).
		Where("rider_id = ?", riderID).
		Select("AVG(rating)").Scan(&avgRating)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items":     list,
			"avgRating": avgRating,
		},
	})
}

// GET /rider/ranking/:type
// getRanking()
func GetRanking(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	rtype := c.Param("type") // income | orders | rating | efficiency

	type Item struct {
		Rank   int     `json:"rank"`
		Name   string  `json:"name"`
		Avatar string  `json:"avatar"`
		Value  float64 `json:"value"`
		IsSelf bool    `json:"isSelf"`
	}

	var items []Item

	switch rtype {
	case "income":
		// 根据总收入排行
		type Row struct {
			RiderID uint
			Name    string
			Avatar  string
			Value   float64
		}
		var rows []Row
		global.Db.Table("rider_wallets").
			Select("rider_wallets.rider_id, rider_profiles.name, rider_profiles.avatar, rider_wallets.total_income AS value").
			Joins("LEFT JOIN rider_profiles ON rider_profiles.rider_id = rider_wallets.rider_id").
			Order("value DESC").
			Limit(50).
			Scan(&rows)

		for i, r := range rows {
			items = append(items, Item{
				Rank:   i + 1,
				Name:   r.Name,
				Avatar: r.Avatar,
				Value:  r.Value,
				IsSelf: r.RiderID == riderID,
			})
		}
	case "orders":
		var profiles []models.RiderProfile
		global.Db.Order("completed_orders DESC").Limit(50).Find(&profiles)

		for i, p := range profiles {
			items = append(items, Item{
				Rank:   i + 1,
				Name:   p.Name,
				Avatar: p.Avatar,
				Value:  float64(p.CompletedOrders),
				IsSelf: p.UserID == riderID,
			})
		}
	case "rating":
		var profiles []models.RiderProfile
		global.Db.Order("rating DESC").Limit(50).Find(&profiles)
		for i, p := range profiles {
			items = append(items, Item{
				Rank:   i + 1,
				Name:   p.Name,
				Avatar: p.Avatar,
				Value:  p.Rating,
				IsSelf: p.UserID == riderID,
			})
		}
	case "efficiency":
		var profiles []models.RiderProfile
		global.Db.Where("online_hours > 0").
			Order("(completed_orders / online_hours) DESC").
			Limit(50).
			Find(&profiles)
		for i, p := range profiles {
			val := 0.0
			if p.OnlineHours > 0 {
				val = float64(p.CompletedOrders) / p.OnlineHours
			}
			items = append(items, Item{
				Rank:   i + 1,
				Name:   p.Name,
				Avatar: p.Avatar,
				Value:  val,
				IsSelf: p.UserID == riderID,
			})
		}
	default:
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "不支持的排行类型"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": items})
}

// ========== 通知 & 系统消息 & 热力图 ==========

// GET /rider/notifications
// getNotifications()
func GetNotifications(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	offset := (page - 1) * size

	readStr := c.Query("read")
	var readFilter *bool
	if readStr != "" {
		v := readStr == "true" || readStr == "1"
		readFilter = &v
	}

	var list []models.RiderNotification
	var total int64

	query := global.Db.Model(&models.RiderNotification{}).
		Where("rider_id = ?", riderID)

	if readFilter != nil {
		query = query.Where("is_read = ?", *readFilter)
	}

	query.Count(&total)
	query.Order("created_at DESC").
		Offset(offset).
		Limit(size).
		Find(&list)

	var unreadCount int64
	global.Db.Model(&models.RiderNotification{}).
		Where("rider_id = ? AND is_read = 0", riderID).
		Count(&unreadCount)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items":       list,
			"unreadCount": unreadCount,
		},
	})
}

// PUT /rider/notifications/:id/read
// markNotificationRead()
func MarkNotificationRead(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	id := c.Param("id")

	if err := global.Db.Model(&models.RiderNotification{}).
		Where("id = ? AND rider_id = ?", id, riderID).
		Update("is_read", true).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}

// GET /rider/messages/system
// getSystemMessages()
func GetSystemMessages(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	offset := (page - 1) * size

	var list []models.RiderSystemMessage
	var total int64

	query := global.Db.Model(&models.RiderSystemMessage{})

	query.Count(&total)
	query.Order("published_at DESC").
		Offset(offset).
		Limit(size).
		Find(&list)

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"items": list,
		},
	})
}

// GET /rider/heatmap
// getHeatmapData()
func GetHeatmapData(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	start := c.Query("start")
	end := c.Query("end")

	var points []models.RiderHeatmapPoint
	query := global.Db.Where("rider_id = ?", riderID)

	if start != "" {
		query = query.Where("date >= ?", start)
	}
	if end != "" {
		query = query.Where("date <= ?", end)
	}
	query.Order("date ASC").Find(&points)

	type Area struct {
		Lat   float64 `json:"lat"`
		Lng   float64 `json:"lng"`
		Count int     `json:"count"`
	}
	type Day struct {
		Date  string `json:"date"`
		Areas []Area `json:"areas"`
	}

	dayMap := map[string][]Area{}
	for _, p := range points {
		key := p.Date.Format("2006-01-02")
		dayMap[key] = append(dayMap[key], Area{
			Lat:   p.Lat,
			Lng:   p.Lng,
			Count: p.Count,
		})
	}

	var result []Day
	for d, areas := range dayMap {
		result = append(result, Day{
			Date:  d,
			Areas: areas,
		})
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": result})
}

// ========== 设置相关 ==========

// GET /rider/settings/work
func GetWorkSettings(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var s models.RiderWorkSettings
	err := global.Db.Where("rider_id = ?", riderID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		// 默认配置
		s = models.RiderWorkSettings{
			RiderID:       riderID,
			AutoAccept:    false,
			DeliveryRange: 5,
			MaxOrders:     5,
		}
		global.Db.Create(&s)
	} else if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	resp := gin.H{
		"autoAccept":    s.AutoAccept,
		"deliveryRange": s.DeliveryRange,
		"workTime": gin.H{
			"start": s.WorkTimeStart,
			"end":   s.WorkTimeEnd,
		},
		"restTime": gin.H{
			"enabled": s.RestEnabled,
			"start":   s.RestStart,
			"end":     s.RestEnd,
		},
		"maxOrders": s.MaxOrders,
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": resp})
}

// PUT /rider/settings/work
func UpdateWorkSettings(c *gin.Context) {
	riderID := c.GetUint("baseUserID")
	var req struct {
		AutoAccept    *bool `json:"autoAccept"`
		DeliveryRange *int  `json:"deliveryRange"`
		WorkTime      *struct {
			Start string `json:"start"`
			End   string `json:"end"`
		} `json:"workTime"`
		RestTime *struct {
			Enabled bool   `json:"enabled"`
			Start   string `json:"start"`
			End     string `json:"end"`
		} `json:"restTime"`
		MaxOrders *int `json:"maxOrders"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	var s models.RiderWorkSettings
	if err := global.Db.Where("rider_id = ?", riderID).First(&s).Error; err == gorm.ErrRecordNotFound {
		s.RiderID = riderID
		global.Db.Create(&s)
	}

	updates := map[string]interface{}{}
	if req.AutoAccept != nil {
		updates["auto_accept"] = *req.AutoAccept
	}
	if req.DeliveryRange != nil {
		updates["delivery_range"] = *req.DeliveryRange
	}
	if req.WorkTime != nil {
		updates["work_time_start"] = req.WorkTime.Start
		updates["work_time_end"] = req.WorkTime.End
	}
	if req.RestTime != nil {
		updates["rest_enabled"] = req.RestTime.Enabled
		updates["rest_start"] = req.RestTime.Start
		updates["rest_end"] = req.RestTime.End
	}
	if req.MaxOrders != nil {
		updates["max_orders"] = *req.MaxOrders
	}

	if len(updates) > 0 {
		if err := global.Db.Model(&models.RiderWorkSettings{}).
			Where("rider_id = ?", riderID).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}

// GET /rider/settings/account
func GetAccountSettings(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var s models.RiderAccountSettings
	err := global.Db.Where("rider_id = ?", riderID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = models.RiderAccountSettings{RiderID: riderID}
		global.Db.Create(&s)
	} else if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"phone":    s.Phone,
			"email":    s.Email,
			"wechat":   s.Wechat,
			"alipay":   s.Alipay,
			"bankCard": s.BankCard,
		},
	})
}

// PUT /rider/settings/account
func UpdateAccountSettings(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var req struct {
		Phone    *string `json:"phone"`
		Email    *string `json:"email"`
		Wechat   *string `json:"wechat"`
		Alipay   *string `json:"alipay"`
		BankCard *string `json:"bankCard"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	var s models.RiderAccountSettings
	if err := global.Db.Where("rider_id = ?", riderID).First(&s).Error; err == gorm.ErrRecordNotFound {
		s.RiderID = riderID
		global.Db.Create(&s)
	}

	updates := map[string]interface{}{}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Wechat != nil {
		updates["wechat"] = *req.Wechat
	}
	if req.Alipay != nil {
		updates["alipay"] = *req.Alipay
	}
	if req.BankCard != nil {
		updates["bank_card"] = *req.BankCard
	}

	if len(updates) > 0 {
		if err := global.Db.Model(&models.RiderAccountSettings{}).
			Where("rider_id = ?", riderID).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}

// GET /rider/settings/notification
func GetNotificationSettings(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var s models.RiderNotificationSettings
	err := global.Db.Where("rider_id = ?", riderID).First(&s).Error
	if err == gorm.ErrRecordNotFound {
		s = models.RiderNotificationSettings{
			RiderID:            riderID,
			OrderNotification:  true,
			SystemNotification: true,
			SoundEnabled:       true,
			VibrationEnabled:   true,
		}
		global.Db.Create(&s)
	} else if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"orderNotification":  s.OrderNotification,
			"systemNotification": s.SystemNotification,
			"soundEnabled":       s.SoundEnabled,
			"vibrationEnabled":   s.VibrationEnabled,
		},
	})
}

// PUT /rider/settings/notification
func UpdateNotificationSettings(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var req struct {
		OrderNotification  *bool `json:"orderNotification"`
		SystemNotification *bool `json:"systemNotification"`
		SoundEnabled       *bool `json:"soundEnabled"`
		VibrationEnabled   *bool `json:"vibrationEnabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	var s models.RiderNotificationSettings
	if err := global.Db.Where("rider_id = ?", riderID).First(&s).Error; err == gorm.ErrRecordNotFound {
		s.RiderID = riderID
		global.Db.Create(&s)
	}

	updates := map[string]interface{}{}
	if req.OrderNotification != nil {
		updates["order_notification"] = *req.OrderNotification
	}
	if req.SystemNotification != nil {
		updates["system_notification"] = *req.SystemNotification
	}
	if req.SoundEnabled != nil {
		updates["sound_enabled"] = *req.SoundEnabled
	}
	if req.VibrationEnabled != nil {
		updates["vibration_enabled"] = *req.VibrationEnabled
	}

	if len(updates) > 0 {
		if err := global.Db.Model(&models.RiderNotificationSettings{}).
			Where("rider_id = ?", riderID).
			Updates(updates).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新失败"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
}

// ========== 认证相关 ==========

// GET /rider/verification
func GetVerification(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var v models.RiderVerification
	err := global.Db.Where("rider_id = ?", riderID).First(&v).Error
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{}})
		return
	} else if err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": v})
}

// POST /rider/verification
func SubmitVerification(c *gin.Context) {
	riderID := c.GetUint("baseUserID")

	var req struct {
		RealName    string `json:"realName"`
		IdCard      string `json:"idCard"`
		IdCardFront string `json:"idCardFront"`
		IdCardBack  string `json:"idCardBack"`
		HealthCert  string `json:"healthCert"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "参数错误"})
		return
	}

	now := time.Now()
	var v models.RiderVerification
	err := global.Db.Where("rider_id = ?", riderID).First(&v).Error
	if err == gorm.ErrRecordNotFound {
		v = models.RiderVerification{
			RiderID:     riderID,
			RealName:    req.RealName,
			IdCard:      req.IdCard,
			IdCardFront: req.IdCardFront,
			IdCardBack:  req.IdCardBack,
			HealthCert:  req.HealthCert,
			Status:      "pending",
			SubmitTime:  now,
		}
		if err := global.Db.Create(&v).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "保存失败"})
			return
		}
	} else if err == nil {
		v.RealName = req.RealName
		v.IdCard = req.IdCard
		v.IdCardFront = req.IdCardFront
		v.IdCardBack = req.IdCardBack
		v.HealthCert = req.HealthCert
		v.Status = "pending"
		v.SubmitTime = now
		if err := global.Db.Save(&v).Error; err != nil {
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "更新失败"})
			return
		}
	} else {
		c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 1,
		"data": gin.H{
			"success":        true,
			"verificationId": v.ID,
		},
	})
}
