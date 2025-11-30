package utils

import (
	"backend/global"
	"backend/models"
	"fmt"
	"time"
	// "gorm.io/gorm"
)

// cleanupPendingLoop 定期清理已过期的 pending 订单（payinfo.ExpiresAt < now）
func StartPendingCleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		<-ticker.C
		now := time.Now()
		var pays []models.PayInfo
		if err := global.Db.Where("status = ? AND expires_at IS NOT NULL AND expires_at < ?", "pending", now).Find(&pays).Error; err != nil {
			fmt.Println("pending cleanup: find payinfo failed:", err)
			continue
		}
		for _, p := range pays {
			// 对每个 payinfo，在事务内删除关联的 pending orders（status=0）及其 order_dishes/order_meals
			tx := global.Db.Begin()
			var orders []models.Order
			if err := tx.Where("pay_infoid = ? AND status = ?", p.ID, 0).Find(&orders).Error; err != nil {
				tx.Rollback()
				fmt.Println("pending cleanup: find orders failed:", err)
				continue
			}
			for _, o := range orders {
				if err := tx.Where("order_id = ?", o.ID).Delete(&models.OrderDish{}).Error; err != nil {
					tx.Rollback()
					fmt.Println("pending cleanup: delete order_dishes failed:", err)
					break
				}
				if err := tx.Where("order_id = ?", o.ID).Delete(&models.OrderMeal{}).Error; err != nil {
					tx.Rollback()
					fmt.Println("pending cleanup: delete order_meals failed:", err)
					break
				}
				if err := tx.Delete(&models.Order{}, o.ID).Error; err != nil {
					tx.Rollback()
					fmt.Println("pending cleanup: delete order failed:", err)
					break
				}
			}
			// 若所有 order 都删除成功，删除 payinfo
			if err := tx.Delete(&models.PayInfo{}, p.ID).Error; err != nil {
				tx.Rollback()
				fmt.Println("pending cleanup: delete payinfo failed:", err)
				continue
			}
			if err := tx.Commit().Error; err != nil {
				fmt.Println("pending cleanup: commit failed:", err)
			} else {
				fmt.Printf("pending cleanup: removed payinfo %d and its expired orders\n", p.ID)
			}
		}
	}
}
