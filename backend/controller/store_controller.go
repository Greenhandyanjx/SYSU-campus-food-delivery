package controller

import (
	"backend/global"
	"backend/models"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetStores 返回商家列表及其若干菜品，供用户首页展示（无需鉴权）
func GetStores(c *gin.Context) {
	var merchants []models.Merchant
	if err := global.Db.Find(&merchants).Error; err != nil {
		log.Printf("GetStores: failed to query merchants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "query merchants failed", "err": err.Error()})
		return
	}

	log.Printf("GetStores: merchants found=%d", len(merchants))

	stores := make([]gin.H, 0, len(merchants))
	for _, m := range merchants {
		log.Printf("GetStores: processing merchant: id=%d base_id=%d shop_name=%s", m.ID, m.BaseID, m.ShopName)
		// Dish.MerchantID 存储的是 base_user id（即 merchants.base_id）
		var dishes []models.Dish
		if err := global.Db.Where("merchant_id = ? AND status = 1", m.BaseID).Limit(6).Find(&dishes).Error; err != nil {
			// 记录具体错误以便调试，但仍继续返回其他商家
			log.Printf("GetStores: failed to query dishes for merchant base_id=%d: %v", m.BaseID, err)
			dishes = []models.Dish{}
		}
		dishItems := make([]gin.H, 0, len(dishes))
		for _, d := range dishes {
			dishItems = append(dishItems, gin.H{
				"id":         d.ID,
				"name":       d.DishName,
				"price":      d.Price,
				"image":      d.ImagePath,
				"categoryId": d.Category,
			})
		}

		stores = append(stores, gin.H{
			"id":          m.ID,
			"name":        m.ShopName,
			"desc":        m.ShopLocation,
			"logo":        m.Logo,
			"tags":        []string{},
			"rating":      4.8,
			"sales":       m.MenuCount,
			"minOrder":    0,
			"deliveryFee": 0,
			"dishes":      dishItems,
		})
	}

	log.Printf("GetStores: returning stores=%d", len(stores))
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": stores})
}
