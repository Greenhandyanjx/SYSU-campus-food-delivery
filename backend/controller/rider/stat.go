package rider

import (
	"backend/global"
	"backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

type RiderStatResp struct {
	NewCount          int64   `json:"newCount"`
	OngoingCount      int64   `json:"ongoingCount"`
	HistoryCount      int64   `json:"historyCount"`
	CompletedOrders   int     `json:"completedOrders"`
	TotalIncome       float64 `json:"totalIncome"`
	AvgDeliverMinutes int     `json:"avgDeliverMinutes"`
}

// GET /api/rider/stat
func GetStat(c *gin.Context) {
	baseUserID := c.GetUint("baseUserID")

	// Rider.ID（orders.rider_id 用这个）
	riderID, err := getRiderIDFromBaseUser(baseUserID)
	if err != nil {
		fail(c, "未找到骑手身份（Rider 表）")
		return
	}

	var newCnt int64
	var ongoingCnt int64
	var historyCnt int64

	// 1) 待接单：全站新订单（status=New 且 rider_id=0）
	global.Db.Table("orders").
		Where("status = ? AND rider_id = 0", OrderStatusToDeliver).
		Count(&newCnt)

	// 2) 进行中：自己的 3/4
	global.Db.Table("orders").
		Where("rider_id = ? AND status IN ?", riderID, []int{OrderStatusToDeliver, OrderStatusDelivering}).
		Count(&ongoingCnt)

	// 3) 历史：自己的 5
	global.Db.Table("orders").
		Where("rider_id = ? AND status = ?", riderID, OrderStatusDone).
		Count(&historyCnt)

	// 4) completedOrders：从 RiderProfile（你已经在结算里 +1）
	var p models.RiderProfile
	completed := 0
	if err := global.Db.Where("user_id = ?", baseUserID).First(&p).Error; err == nil {
		completed = p.CompletedOrders
	} else {
		// 找不到 profile 就退化成 historyCnt
		completed = int(historyCnt)
	}

	// 5) totalIncome：优先走钱包（最快）
	totalIncome := 0.0
	var w models.RiderWallet
	if err := global.Db.Where("rider_id = ?", riderID).First(&w).Error; err == nil {
		totalIncome = w.TotalIncome
	}

	// 6) avgDeliverMinutes：平均 (finish_at - accepted_at)
	// MySQL: TIMESTAMPDIFF(MINUTE, accepted_at, finish_at)
	type avgRow struct{ Avg float64 }
	var ar avgRow
	global.Db.Raw(`
		SELECT COALESCE(AVG(TIMESTAMPDIFF(MINUTE, accepted_at, finish_at)), 0) AS avg
		FROM orders
		WHERE rider_id = ? AND status = ? AND accepted_at IS NOT NULL AND finish_at IS NOT NULL
	`, riderID, OrderStatusDone).Scan(&ar)

	avgMin := int(ar.Avg + 0.5) // 四舍五入

	ok(c, RiderStatResp{
		NewCount:          newCnt,
		OngoingCount:      ongoingCnt,
		HistoryCount:      historyCnt,
		CompletedOrders:   completed,
		TotalIncome:       totalIncome,
		AvgDeliverMinutes: avgMin,
	})
}

// 可选：给你一个“今天收入”的字段（要就把它加回去）
func _todayRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	end := start.Add(24 * time.Hour)
	return start, end
}
