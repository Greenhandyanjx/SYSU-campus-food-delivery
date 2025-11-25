package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"errors"
	"fmt"
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

// 添加到购物车功能
func AddToCart(c *gin.Context) {
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Fail(c, "参数错误："+err.Error())
		return
	}

	userID := c.MustGet("baseUserID").(uint)
	fmt.Println("UserID:", userID, "Request:", req)

	// 从 map 中读取前端实际传的字段
	merchantIDFloat, ok := req["storeId"].(float64)
	if !ok {
		utils.Fail(c, "storeId 参数错误")
		return
	}
	dishIDFloat, ok := req["dishId"].(float64)
	if !ok {
		utils.Fail(c, "dishId 参数错误")
		return
	}
	qtyFloat, ok := req["qty"].(float64)
	if !ok {
		utils.Fail(c, "qty 参数错误")
		return
	}

	merchantID := uint(merchantIDFloat)
	dishID := uint(dishIDFloat)
	qty := int(qtyFloat)

	// 1. 找用户的购物车
	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		fmt.Println("Cart not found, creating new one")
		cart = models.Cart{UserID: userID}
		if err := global.Db.Create(&cart).Error; err != nil {
			fmt.Println("Create cart failed:", err)
			utils.Error(c, err)
			return
		}
		fmt.Println("New cart created:", cart)
	} else {
		fmt.Println("Found cart:", cart)
	}

	// 2. 是否已经存在同商家 + 同菜品？
	var item models.CartItem
	err := global.Db.Where(
		"cart_id = ? AND merchant_id = ? AND dish_id = ?",
		cart.ID, merchantID, dishID,
	).First(&item).Error

	if err == nil {
		fmt.Println("CartItem exists, increasing quantity:", item)
		item.Qty += qty
		if err := global.Db.Save(&item).Error; err != nil {
			fmt.Println("Failed to update item:", err)
			utils.Error(c, err)
			return
		}
		fmt.Println("Updated item:", item)
		utils.Success(c, "添加成功（数量增加）")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("Error querying cart_item:", err)
		utils.Error(c, err)
		return
	}

	// 3. 获取菜品信息（可选，如果前端传了 name 和 price 可直接用）
	var dish models.Dish
	if err := global.Db.Where("id = ?", dishID).First(&dish).Error; err != nil {
		fmt.Println("Dish not found, using frontend data")
		dish.DishName, _ = req["name"].(string)
		dish.Price, _ = req["price"].(string)
	} else {
		fmt.Println("Found dish in DB:", dish)
	}

	// 4. 新增购物车 item
	newItem := models.CartItem{
		CartID:     cart.ID,
		MerchantID: merchantID,
		DishID:     dishID,
		Name:       dish.DishName,
		Price:      dish.Price, // 已经是 string
		Qty:        qty,
		Selected:   true,
	}

	fmt.Println("Inserting new CartItem:", newItem)
	if err := global.Db.Create(&newItem).Error; err != nil {
		fmt.Println("Create CartItem failed:", err)
		utils.Error(c, err)
		return
	}

	fmt.Println("Create CartItem success")
	utils.Success(c, "添加成功")
}

// 更新用户购物车
func UpdateCartItem(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)
	// 解析请求参数
	var req struct {
		StoreID string `json:"storeId" binding:"required"`
		DishID  int    `json:"dishId" binding:"required"`
		Qty     int    `json:"qty" binding:"required,min=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, err)
		return
	}
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

	// 转换 storeID 为 uint
	merchantID, err := strconv.ParseUint(req.StoreID, 10, 32)
	if err != nil {
		utils.Error(c, err)
		return
	}

	// 处理购物车商品
	if req.Qty == 0 {
		// 如果数量为0，删除该商品
		if err := global.Db.Where("cart_id = ? AND dish_id = ? AND merchant_id = ?", cart.ID, uint(req.DishID), uint(merchantID)).Delete(&models.CartItem{}).Error; err != nil {
			utils.Error(c, err)
			return
		}
	} else {
		// 更新或创建购物车商品
		var cartItem models.CartItem
		if err := global.Db.Where("cart_id = ? AND dish_id = ? AND merchant_id = ?", cart.ID, uint(req.DishID), uint(merchantID)).First(&cartItem).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// 创建新商品项
				cartItem = models.CartItem{
					CartID:     cart.ID,
					DishID:     uint(req.DishID),
					MerchantID: uint(merchantID),
					Qty:        req.Qty,
				}
				if err := global.Db.Create(&cartItem).Error; err != nil {
					utils.Error(c, err)
					return
				}
			} else {
				utils.Error(c, err)
				return
			}
		} else {
			// 更新现有商品数量
			if err := global.Db.Model(&cartItem).Update("qty", req.Qty).Error; err != nil {
				utils.Error(c, err)
				return
			}
		}
	}

	// 返回成功响应
	utils.Success(c, gin.H{
		"success": true,
		"updatedItem": gin.H{
			"storeId": req.StoreID,
			"dishId":  req.DishID,
			"qty":     req.Qty,
		},
	})
}

// SelectItem - 单个商品选中/取消
func SelectItem(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)

	var req struct {
		ItemID   uint `json:"dishId" binding:"required"`   // 前端传 dishId
		Selected bool `json:"selected" binding:"required"` // true=选中, false=取消
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, err)
		return
	}

	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		utils.Error(c, err)
		return
	}

	// 布尔转 int8
	selected := int8(0)
	if req.Selected {
		selected = 1
	}

	if err := global.Db.Model(&models.CartItem{}).
		Where("cart_id = ? AND dish_id = ?", cart.ID, req.ItemID).
		Update("selected", selected).Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"dishId":   req.ItemID,
		"selected": req.Selected,
	})
}

// SelectShop - 店铺内全部商品选中/取消
func SelectShop(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)

	var req struct {
		StoreID  uint `json:"storeId" binding:"required"`
		Selected bool `json:"selected" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, err)
		return
	}

	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		utils.Error(c, err)
		return
	}

	selected := int8(0)
	if req.Selected {
		selected = 1
	}

	if err := global.Db.Model(&models.CartItem{}).
		Where("cart_id = ? AND merchant_id = ?", cart.ID, req.StoreID).
		Update("selected", selected).Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"storeId":  req.StoreID,
		"selected": req.Selected,
	})
}

// SelectAll - 购物车所有商品选中/取消
func SelectAll(c *gin.Context) {
	userID := c.MustGet("baseUserID").(uint)

	var req struct {
		Selected bool `json:"selected" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c, err)
		return
	}

	var cart models.Cart
	if err := global.Db.Where("user_id = ?", userID).First(&cart).Error; err != nil {
		utils.Error(c, err)
		return
	}

	selected := int8(0)
	if req.Selected {
		selected = 1
	}

	if err := global.Db.Model(&models.CartItem{}).
		Where("cart_id = ?", cart.ID).
		Update("selected", selected).Error; err != nil {
		utils.Error(c, err)
		return
	}

	utils.Success(c, gin.H{
		"selected": req.Selected,
	})
}
