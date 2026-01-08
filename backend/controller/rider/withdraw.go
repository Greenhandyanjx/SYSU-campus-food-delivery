package rider

import (
	"backend/global"
	"backend/models"
	"time"

	"github.com/gin-gonic/gin"
)

func ApplyWithdraw(c *gin.Context) {
	riderID, okk := getRiderIDByBase(c)
	if !okk {
		return
	}

	var req struct {
		Amount  float64 `json:"amount"`
		Account string  `json:"account"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Amount <= 0 || req.Account == "" {
		fail(c, "参数错误")
		return
	}

	tx := global.Db.Begin()

	// 锁钱包行，避免并发扣款
	var w models.RiderWallet
	if err := tx.Where("rider_id = ?", riderID).First(&w).Error; err != nil {
		w = models.RiderWallet{RiderID: riderID}
		if err2 := tx.Create(&w).Error; err2 != nil {
			tx.Rollback()
			fail(c, "钱包不存在且创建失败")
			return
		}
	}

	if w.Balance < req.Amount {
		tx.Rollback()
		fail(c, "余额不足")
		return
	}

	// 余额转冻结
	if err := tx.Model(&models.RiderWallet{}).Where("id = ?", w.ID).
		Updates(map[string]any{
			"balance":       w.Balance - req.Amount,
			"frozen_amount": w.FrozenAmount + req.Amount,
		}).Error; err != nil {
		tx.Rollback()
		fail(c, "更新钱包失败")
		return
	}

	wd := models.RiderWithdraw{
		RiderID:   riderID,
		Amount:    req.Amount,
		Account:   req.Account,
		Status:    "pending",
		AppliedAt: time.Now(),
	}
	if err := tx.Create(&wd).Error; err != nil {
		tx.Rollback()
		fail(c, "创建提现失败")
		return
	}

	tx.Commit()
	ok(c, gin.H{"success": true, "withdrawId": wd.ID})
}

func GetWithdraws(c *gin.Context) {
	riderID, okk := getRiderIDByBase(c)
	if !okk {
		return
	}
	var list []models.RiderWithdraw
	global.Db.Where("rider_id = ?", riderID).Order("applied_at DESC").Limit(100).Find(&list)
	ok(c, list)
}
