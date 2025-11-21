package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"

	"github.com/gin-gonic/gin"
)

// GET /user/cart
func GetCart(c *gin.Context) {
	// 假设你的 JWT 中已经写入 user_id
	userID := c.GetString("user_id")
	if userID == "" {
		utils.Fail(c, "未找到用户 ID")
		return
	}

	// 查找用户购物车
	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		utils.Success(c, gin.H{"shops": []interface{}{}})
		return
	}

	// 查找购物车条目（cart_items）
	var items []models.CartItem
	global.Db.Where("cart_id = ?", cart.ID).Find(&items)

	// shopId → shop 数据
	shopMap := map[string]*models.CartShopResponse{}

	for _, item := range items {
		// 初始化 shop
		if _, exists := shopMap[item.StoreID]; !exists {
			shopMap[item.StoreID] = &models.CartShopResponse{
				StoreID: item.StoreID,
				Name:    item.StoreName,
				Items:   []models.CartItemResponse{},
			}
		}

		// 添加菜品
		shopMap[item.StoreID].Items = append(shopMap[item.StoreID].Items, models.CartItemResponse{
			DishID:      item.DishID,
			Name:        item.Name,
			Price:       item.Price,
			Qty:         item.Qty,
			OriginalQty: item.Qty,
			Selected:    item.Selected == true,
			Category:    item.Category,
		})
	}

	// 转为数组 shops:[]
	shops := []models.CartShopResponse{}
	for _, shop := range shopMap {
		shops = append(shops, *shop)
	}

	utils.Success(c, gin.H{"shops": shops})
}
