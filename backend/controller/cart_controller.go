package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AddToCartRequest struct {
	MerchantID uint `json:"merchantId" binding:"required"`
	DishID     uint `json:"dishId" binding:"required"`
	Qty        int  `json:"qty" binding:"required"`
}

// GET /user/cart
func GetUserCart(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	storeID := c.Query("storeId")
	// 查询用户购物车，如果不存在则创建
	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 自动创建空购物车
			cart = models.Cart{UserID: userID}
			if err := global.Db.Create(&cart).Error; err != nil {
				utils.Error(c, err)
				return
			}
		} else {
			utils.Error(c, err)
			return
		}
	}
	// 查询购物车项
	var items []models.CartItem
	query := global.Db.Where("cart_id = ?", cart.ID)
	if storeID != "" {
		merchantID, _ := strconv.ParseUint(storeID, 10, 32)
		query = query.Where("merchant_id = ?", uint(merchantID))
	}
	if err := query.Find(&items).Error; err != nil {
		utils.Error(c, err)
		return
	}
	// 如果指定了商家，直接返回
	if storeID != "" {
		merchantID, _ := strconv.ParseUint(storeID, 10, 32)
		var merchant models.Merchant
		global.Db.Where("base_id = ?", uint(merchantID)).First(&merchant)

		utils.Success(c, gin.H{
			"merchant_id":   storeID,
			"merchant_name": merchant.ShopName, // 使用 shop_name
			"items":         items,
		})
		return
	}
	// 按商家分组
	shopsMap := make(map[uint]gin.H)
	for _, item := range items {
		merchantID := item.MerchantID
		if _, exists := shopsMap[merchantID]; !exists {
			// 查询商家信息
			var merchant models.Merchant
			global.Db.Where("base_id = ?", merchantID).First(&merchant)

			shopsMap[merchantID] = gin.H{
				"merchant_id":   merchantID,
				"merchant_name": merchant.ShopName,
				"shop_location": merchant.ShopLocation,
				"owner":         merchant.Owner,
				"phone":         merchant.Phone,
				"logo":          merchant.Logo,
				"status":        merchant.Status,
				"items":         []models.CartItem{},
			}
		}
		shop := shopsMap[merchantID]
		shop["items"] = append(shop["items"].([]models.CartItem), item)
		shopsMap[merchantID] = shop
	}
	// 转换为数组
	shops := make([]gin.H, 0, len(shopsMap))
	for _, shop := range shopsMap {
		shops = append(shops, shop)
	}

	utils.Success(c, gin.H{
		"shops": shops,
	})
}

// POST /user/cart/add
func AddToCart(c *gin.Context) {
	var req AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数错误："+err.Error())
		return
	}

	userID := c.MustGet("baseUserID").(uint)

	// 1. 找用户的购物车（一个用户只有一个 cart）
	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		// 如果没有购物车，则创建一个
		cart = models.Cart{
			UserID: userID,
		}
		if err := global.Db.Create(&cart).Error; err != nil {
			utils.Error(c, err)
			return
		}
	}

	// 2. 是否已经存在同商家 + 同菜品？
	var item models.CartItem
	err := global.Db.Where(
		"cart_id = ? AND merchant_id = ? AND dish_id = ?",
		cart.ID, req.MerchantID, req.DishID,
	).First(&item).Error

	if err == nil {
		// 已经存在 → 数量 + req.Qty
		item.Qty += req.Qty
		if err := global.Db.Save(&item).Error; err != nil {
			utils.Error(c, err)
			return
		}
		utils.Success(c, "添加成功（数量增加）")
		return
	}

	// 3. 获取菜品信息（用于价格与名称）
	var dish models.Dish
	if err := global.Db.Where("id = ?", req.DishID).First(&dish).Error; err != nil {
		utils.Fail(c, "菜品不存在")
		return
	}

	// 4. 新增购物车 item
	newItem := models.CartItem{
		CartID:     cart.ID,
		MerchantID: req.MerchantID,
		DishID:     req.DishID,
		Name:       dish.DishName,
		Price:      dish.Price,
		Qty:        req.Qty,
		Selected:   true,
	}

	if err := global.Db.Create(&newItem).Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, "添加成功")
}
