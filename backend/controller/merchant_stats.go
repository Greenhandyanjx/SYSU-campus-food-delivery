package controller

import (
	"backend/global"
	"backend/models"
	"fmt"
	"log"
)

// UpdateMerchantTopCategories 统计指定商家的菜品与套餐分类数量（包含 dishes 与 meals），
// 选出出现次数最多的两个分类并更新 merchants 表中的 TopCategory1/TopCategory2 字段。
func UpdateMerchantTopCategories(merchantID uint) {
	if merchantID == 0 {
		return
	}

	// 聚合来自 dishes 与 meals 的分类计数
	// 返回格式：category, cnt
	sql := `
    SELECT category, SUM(cnt) AS total
    FROM (
      SELECT category, COUNT(*) AS cnt FROM dishes WHERE merchant_id = ? GROUP BY category
      UNION ALL
      SELECT category, COUNT(*) AS cnt FROM meals WHERE merchant_id = ? GROUP BY category
    ) t
    GROUP BY category
    ORDER BY total DESC
    LIMIT 2
    `

	type r struct {
		Category int
		Total    int
	}
	var rows []r
	if err := global.Db.Raw(sql, merchantID, merchantID).Scan(&rows).Error; err != nil {
		log.Printf("UpdateMerchantTopCategories: db error: %v", err)
		return
	}

	var top1, top2 int
	if len(rows) > 0 {
		top1 = rows[0].Category
	}
	if len(rows) > 1 {
		top2 = rows[1].Category
	}

	// 更新 merchants 表的 TopCategory1/TopCategory2
	if err := global.Db.Model(&models.Merchant{}).Where("id = ?", merchantID).Updates(map[string]interface{}{"top_category1": top1, "top_category2": top2}).Error; err != nil {
		log.Printf("UpdateMerchantTopCategories: update merchant error: %v", err)
	} else {
		fmt.Printf("Updated merchant %d top categories: %d, %d\n", merchantID, top1, top2)
	}
}
