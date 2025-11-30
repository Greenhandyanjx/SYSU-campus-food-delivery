package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// GetStores 返回商家列表及其若干菜品，供用户首页展示（无需鉴权）
func GetStores(c *gin.Context) {
	ctx := context.Background()
	// try read from cache first
	var cached []map[string]interface{}
	if ok, _ := utils.GetJSON(ctx, "stores:all", &cached); ok {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": cached})
		return
	}

	var merchants []models.Merchant
	// limit number of merchants returned to avoid heavy full-table scans
	// (frontend displays a page of stores; returning too many merchants will slow queries)
	if err := global.Db.Limit(50).Find(&merchants).Error; err != nil {
		log.Printf("GetStores: failed to query merchants: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "msg": "query merchants failed", "err": err.Error()})
		return
	}

	log.Printf("GetStores: merchants found=%d", len(merchants))

	stores := make([]map[string]interface{}, 0, len(merchants))

	// collect base_ids to batch query dishes
	baseIDs := make([]uint, 0, len(merchants))
	for _, m := range merchants {
		baseIDs = append(baseIDs, m.BaseID)
	}

	// batch load dishes for all merchants (status = 1)
	var allDishes []models.Dish
	if len(baseIDs) > 0 {
		if err := global.Db.Where("merchant_id IN ? AND status = 1", baseIDs).Find(&allDishes).Error; err != nil {
			log.Printf("GetStores: failed to batch query dishes: %v", err)
			allDishes = []models.Dish{}
		}
	}

	// group dishes by merchant_id (which stores base_id)
	dishesByBase := make(map[uint][]models.Dish)
	for _, d := range allDishes {
		dishesByBase[d.MerchantID] = append(dishesByBase[d.MerchantID], d)
	}

	// assemble stores with up to 6 dishes each
	for _, m := range merchants {
		ds := dishesByBase[m.BaseID]
		if len(ds) > 6 {
			ds = ds[:6]
		}
		dishItems := make([]map[string]interface{}, 0, len(ds))
		for _, d := range ds {
			dishItems = append(dishItems, map[string]interface{}{
				"id":         d.ID,
				"name":       d.DishName,
				"price":      d.Price,
				"image":      d.ImagePath,
				"categoryId": d.Category,
			})
		}

		stores = append(stores, map[string]interface{}{
			"id":          m.ID,
			"base_id":     m.BaseID,
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

	// cache for 60s asynchronously
	go func(data interface{}) {
		_ = utils.SetJSON(context.Background(), "stores:all", data, 60*time.Second)
	}(stores)

	log.Printf("GetStores: returning stores=%d", len(stores))
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": stores})
}

// GetStoreByQuery 支持通过 id 或 name 或 base_id 查询单个店铺并返回基础信息（带缓存）
func GetStoreByQuery(c *gin.Context) {
	ctx := context.Background()
	// prefer id, then base_id, then name
	id := c.Query("id")
	baseID := c.Query("base_id")
	name := c.Query("name")

	var key string
	if id != "" {
		key = "store:id:" + id
	} else if baseID != "" {
		key = "store:base_id:" + baseID
	} else if name != "" {
		key = "store:name:" + name
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "id or base_id or name required"})
		return
	}

	var cached map[string]interface{}
	if ok, _ := utils.GetJSON(ctx, key, &cached); ok {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": cached})
		return
	}

	var m models.Merchant
	var err error
	if id != "" {
		err = global.Db.First(&m, id).Error
	} else if baseID != "" {
		var bid uint
		// parse uint
		_, err = fmt.Sscan(baseID, &bid)
		if err == nil {
			err = global.Db.Where("base_id = ?", bid).First(&m).Error
		}
	} else {
		err = global.Db.Where("shop_name = ?", name).First(&m).Error
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found", "err": err.Error()})
		return
	}

	resp := map[string]interface{}{
		"id":            m.ID,
		"base_id":       m.BaseID,
		"name":          m.ShopName,
		"desc":          m.ShopLocation,
		"shop_location": m.ShopLocation,
		"logo":          m.Logo,
		"phone":         m.Phone,
		"menu_count":    m.MenuCount,
	}

	// cache for 60s
	go utils.SetJSON(context.Background(), key, resp, 60*time.Second)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": resp})
}

// GetStoreDishes 返回单个店铺的所有菜品与套餐
func GetStoreDishes(c *gin.Context) {
	ctx := context.Background()
	storeId := c.Query("storeId")
	baseId := c.Query("baseId")
	if storeId == "" && baseId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "storeId or baseId required"})
		return
	}

	var key string
	if baseId != "" {
		key = "store:data:base_id:" + baseId
	} else {
		key = "store:data:id:" + storeId
	}

	var cached map[string]interface{}
	if ok, _ := utils.GetJSON(ctx, key, &cached); ok {
		c.JSON(http.StatusOK, gin.H{"code": 1, "data": cached})
		return
	}

	// Find merchant to determine base_id
	var m models.Merchant
	var err error
	if baseId != "" {
		var bid uint
		_, err = fmt.Sscan(baseId, &bid)
		if err == nil {
			err = global.Db.Where("base_id = ?", bid).First(&m).Error
		}
	} else {
		err = global.Db.First(&m, storeId).Error
	}
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "msg": "merchant not found", "err": err.Error()})
		return
	}

	// get dishes and meals by merchant's base id (Dish.MerchantID stores base user id)
	var dishes []models.Dish
	var meals []models.Meal
	if err := global.Db.Where("merchant_id = ?", m.BaseID).Find(&dishes).Error; err != nil {
		log.Printf("GetStoreDishes: dishes query failed for base_id=%d: %v", m.BaseID, err)
		dishes = []models.Dish{}
	}
	if err := global.Db.Where("merchant_id = ?", m.BaseID).Find(&meals).Error; err != nil {
		log.Printf("GetStoreDishes: meals query failed for base_id=%d: %v", m.BaseID, err)
		meals = []models.Meal{}
	}

	resp := map[string]interface{}{
		"merchant": map[string]interface{}{"id": m.ID, "base_id": m.BaseID, "name": m.ShopName, "logo": m.Logo, "phone": m.Phone, "shop_location": m.ShopLocation},
		"dishes":   dishes,
		"meals":    meals,
	}

	go utils.SetJSON(context.Background(), key, resp, 60*time.Second)

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": resp})
}
