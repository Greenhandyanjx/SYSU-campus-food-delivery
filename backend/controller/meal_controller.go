package controller

import (
	"backend/global"
	"backend/models"
	"backend/utils"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Meal_add(ctx *gin.Context) {
	var meal models.Meal
	// 中间结构体用于绑定
	type SetmealDish struct {
		Copies int    `json:"copies"`
		Name   string `json:"name"`
		Price  string `json:"price"`
	}
	type mealRequest struct {
		DishIDs []SetmealDish `json:"setmealDishes"`
	}
	var request mealRequest
	//
	baseUserID := ctx.MustGet("baseUserID").(uint)
	// 将用户ID赋给套餐的MerchantID字段
	meal.MerchantID = baseUserID
	// 验证 category 在允许范围内（1..15）
	//第一次绑定
	body, _ := io.ReadAll(ctx.Request.Body)
	fmt.Println("Request Body:", string(body))             // 打印请求体内容
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body)) // 重置请求体
	if err := ctx.ShouldBindJSON(&meal); err != nil {
		// 打印绑定错误的详细信息
		log.Printf("绑定错误: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "binding error",
		})
		return
	}
	if meal.Category < 1 || meal.Category > 15 {
		log.Printf("无效的类别ID: %d", meal.Category)
		ctx.JSON(http.StatusBadRequest, gin.H{"code": 0, "msg": "invalid category id"})
		return
	}

	fmt.Println("Request Body:", string(body)) // 打印请求体内容
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(body))
	// 获取 dish_ids，第二次绑定
	if err := ctx.ShouldBindJSON(&request); err != nil {
		// 打印绑定错误的详细信息
		log.Printf("绑定错误: %v", err.Error())
		ctx.JSON(http.StatusBadRequest, gin.H{
			"code": "400",
			"msg":  "binding dish_ids error",
		})
		return
	}

	if err := global.Db.Create(&meal).Error; err != nil {
		fmt.Println(err.Error())
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误，请稍后再试",
		})
		return
	}

	// 创建 meal_dish 关联关系
	for _, setmealDish := range request.DishIDs {
		mealDish := models.MealDish{
			MealID: meal.ID,
			DishID: 0, // 假设 setmealDish 中没有 DishID，你需要从 DishList 或其他地方获取
			Num:    setmealDish.Copies,
		}
		// 如果 setmealDish 中有 DishID，可以直接使用
		// 如果没有，你需要根据 setmealDish.Name 或其他信息查询 DishID
		// 这里假设你需要通过名称查询 DishID
		var dish models.Dish
		if err := global.Db.Where("dish_name = ?", setmealDish.Name).First(&dish).Error; err != nil {
			log.Printf("查询 DishID 错误: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": "500",
				"msg":  "查询 DishID 错误",
			})
			return
		}
		mealDish.DishID = dish.ID
		if err := global.Db.Create(&mealDish).Error; err != nil {
			log.Printf("创建关联关系错误: %v", err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"code": "500",
				"msg":  "服务器内部错误，请稍后再试",
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"code": "200",
		"msg":  "套餐创建成功",
	})
	// 创建成功后更新商家分类统计
	go UpdateMerchantTopCategories(meal.MerchantID)
	// 清除用户首页与该商家详情缓存
	go func(mid uint) {
		_ = utils.Del(context.Background(), "stores:all")
		_ = utils.Del(context.Background(), fmt.Sprintf("store:data:base_id:%d", mid))
		_ = utils.Del(context.Background(), fmt.Sprintf("store:base_id:%d", mid))
	}(meal.MerchantID)
}

func Meal_Edit(c *gin.Context) {
	var meal models.Meal
	if err := c.ShouldBindJSON(&meal); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
		return
	}

	// 根据 ID 查找套餐
	var existingMeal models.Meal
	if err := global.Db.First(&existingMeal, meal.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "套餐未找到", "data": nil})
		return
	}

	// 更新套餐信息
	if err := global.Db.Model(&existingMeal).Updates(models.Meal{
		Mealname:    meal.Mealname,
		Price:       meal.Price,
		Description: meal.Description,
		MerchantID:  meal.MerchantID,
		Status:      meal.Status,
		ImagePath:   meal.ImagePath,
		Category:    meal.Category,
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新套餐信息失败", "data": nil})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true, "mealId": strconv.Itoa(meal.ID)}})
	// 编辑成功后更新商家分类统计
	go UpdateMerchantTopCategories(existingMeal.MerchantID)
	// 清除相关缓存
	go func(mid uint) {
		_ = utils.Del(context.Background(), "stores:all")
		_ = utils.Del(context.Background(), fmt.Sprintf("store:data:base_id:%d", mid))
		_ = utils.Del(context.Background(), fmt.Sprintf("store:base_id:%d", mid))
	}(existingMeal.MerchantID)
}

func Meal_Delete(c *gin.Context) {
	// 绑定请求体到 map 结构体
	var request map[string]interface{}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
		return
	}
	// 获取请求中的 id 或 id 列表
	idOrList, ok := request["id"]
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体中缺少 id 字段", "data": nil})
		return
	}
	var removedIDs []int
	switch ids := idOrList.(type) {
	case float64:
		removedIDs = append(removedIDs, int(ids))
		if err := global.Db.Table("meals").Delete(&models.Meal{}, ids).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "删除套餐失败", "data": nil})
			return
		}
	case string:
		// 处理逗号分隔的 ID 字符串
		idStrings := strings.Split(ids, ",")
		for _, idStr := range idStrings {
			idInt, err := strconv.Atoi(idStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "id 字段包含无效数字", "data": nil})
				return
			}
			removedIDs = append(removedIDs, idInt)
			if err := global.Db.Table("meals").Where("id = ?", idInt).Delete(&models.Meal{}).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "删除套餐失败", "data": nil})
				return
			}
		}
	case []interface{}:
		// 批量删除菜品
		for _, id := range ids {
			if idFloat, ok := id.(float64); ok {
				removedIDs = append(removedIDs, int(idFloat))
				if err := global.Db.Table("meals").Delete(&models.Meal{}, int(idFloat)).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "删除套餐失败", "data": nil})
					return
				}
			} else {
				c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "id 列表中包含float64以外类型", "data": nil})
				return
			}
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "id 字段类型错误", "data": nil})
		return
	}
	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"succe ss": true, "removed": removedIDs}})
	// 删除后建议更新商家分类统计（如果可得 merchantID）
	// 清除全量缓存以保证用户端能尽快看到变更
	go func() {
		_ = utils.Del(context.Background(), "stores:all")
	}()
}

func Edit_Meal_Status(c *gin.Context) {
	// 绑定请求体到结构体
	var request struct {
		ID     int    `json:"id" form:"id"`
		Status string `json:"status" form:"status"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求体格式错误", "data": nil})
		return
	}
	// 根据前端传递的 status 字段转换为数据库中的整数值
	var status int
	switch request.Status {
	case "on":
		status = 1
	case "off":
		status = 0
	case "recommended":
		status = 1 // 假设推荐状态对应的值为 2
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "状态值无效", "data": nil})
		return
	}
	// 根据 ID 查找，可忽略
	var existingMeal models.Meal
	if err := global.Db.Table("meals").First(&existingMeal, request.ID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "套餐未找到", "data": nil})
		return
	}
	// 更新菜品状态
	if err := global.Db.Table("meals").Where("id = ?", request.ID).Update("status", status).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "更新套餐状态失败", "data": nil})
		return
	}
	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": gin.H{"success": true}})
	// 清除相关缓存（尝试获取套餐的 merchant 后再失效）
	go func() {
		var meal models.Meal
		if err := global.Db.First(&meal, request.ID).Error; err == nil {
			_ = utils.Del(context.Background(), "stores:all")
			_ = utils.Del(context.Background(), fmt.Sprintf("store:data:base_id:%d", meal.MerchantID))
			_ = utils.Del(context.Background(), fmt.Sprintf("store:base_id:%d", meal.MerchantID))
		} else {
			_ = utils.Del(context.Background(), "stores:all")
		}
	}()
}

// 分页获取套餐信息
func GetMealsPage(c *gin.Context) {
	// 获取请求参数
	page, err1 := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, err2 := strconv.Atoi(c.DefaultQuery("size", "20"))
	name := c.Query("name")
	statusParam := c.Query("status")
	categoryIdParam := c.Query("categoryId") // 获取 categoryId 参数
	// 检查参数是否合法
	if err1 != nil || err2 != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "无效的分页参数", "data": nil})
		return
	}
	// 将 status 参数转换为整数
	var status int
	if statusParam != "" {
		var err error
		status, err = strconv.Atoi(statusParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "无效的状态参数", "data": nil})
			return
		}
	}
	// 将 categoryId 参数转换为整数
	var categoryId int
	if categoryIdParam != "" {
		var err error
		categoryId, err = strconv.Atoi(categoryIdParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "无效的类别参数", "data": nil})
			return
		}
	}
	// 获取上下文中的 baseUserID
	baseUserID, exists := c.Get("baseUserID")
	fmt.Println("id", baseUserID)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": "401",
			"msg":  "未找到商户ID",
		})
		return
	}
	// 确保 baseUserID 是 uint 类型
	merchantID, ok := baseUserID.(uint)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code": "401",
			"msg":  "商户ID类型错误",
		})
		return
	}
	// 计算分页参数
	offset := (page - 1) * size
	// 构建查询条件的字符串，用于缓存的key
	queryConditions := fmt.Sprintf("merchant_id=%d&page=%d&size=%d&name=%s&status=%d&category_id=%d",
		merchantID, page, size, name, status, categoryId)
	// 尝试从 Redis 获取缓存的数据
	var cachedData struct {
		Items []models.Meal
		Total int64
	}
	found, err := utils.GetJSON(context.Background(), queryConditions, &cachedData)
	if err != nil {
		log.Printf("Redis读取错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误",
		})
		return
	}
	if found {
		// 如果成功从Redis获取缓存数据，则直接返回
		items := make([]gin.H, len(cachedData.Items))
		for i, meal := range cachedData.Items {
			items[i] = gin.H{
				"id":         meal.ID,
				"name":       meal.Mealname,
				"price":      meal.Price,
				"status":     meal.Status,
				"image":      meal.ImagePath,
				"categoryId": meal.Category,
			}
		}
		c.JSON(http.StatusOK, gin.H{
			"code": "1",
			"msg":  "获取套餐列表成功",
			"data": gin.H{
				"items": items,
				"total": cachedData.Total,
			},
		})
		return
	}
	// 构建查询条件
	var query = global.Db.Model(&models.Meal{}).Where("merchant_id = ?", merchantID)
	if name != "" {
		query = query.Where("mealname LIKE ?", "%"+name+"%")
	}
	if statusParam != "" {
		query = query.Where("status = ?", status)
	}
	if categoryIdParam != "" {
		query = query.Where("category = ?", categoryId)
	}
	// 获取总记录数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		log.Printf("获取套餐总数失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "获取套餐总数失败", "data": nil})
		return
	}
	// 获取分页数据
	var meals []models.Meal
	if err := query.Offset(offset).Limit(size).Find(&meals).Error; err != nil {
		log.Printf("获取套餐列表失败: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "获取套餐列表失败", "data": nil})
		return
	}
	// 准备返回数据
	items := make([]gin.H, len(meals))
	for i, meal := range meals {
		items[i] = gin.H{
			"id":         meal.ID,
			"name":       meal.Mealname,
			"price":      meal.Price,
			"status":     meal.Status,
			"image":      meal.ImagePath,
			"categoryId": meal.Category,
			"stock":      0, // 假设 stock 字段在 Meal 结构体中不存在，这里返回 0
		}
	}
	// 序列化数据并存入Redis
	cachedData = struct {
		Items []models.Meal
		Total int64
	}{
		Items: meals,
		Total: total,
	}
	err = utils.SetJSON(context.Background(), queryConditions, cachedData, 5*time.Minute)
	if err != nil {
		log.Printf("Redis写入错误: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": "500",
			"msg":  "服务器内部错误",
		})
		return
	}
	// 返回结果
	c.JSON(http.StatusOK, gin.H{
		"code": "1",
		"msg":  "获取套餐列表成功",
		"data": gin.H{
			"items": items,
			"total": total,
		},
	})
}

func Get_Meal_ById(c *gin.Context) {
	// 获取请求参数中的 id
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "请求参数中缺少 id 字段", "data": nil})
		return
	}
	// 构建查询条件
	var meal models.Meal
	if err := global.Db.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "套餐不存在", "data": nil})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "查询套餐失败", "data": nil})
		}
		return
	}
	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"code": 1, "data": meal})
}

// GetPublicMealById 不需要鉴权，返回单个 meal 及其 setmealDishes，供用户预览使用
func GetPublicMealById(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 0, "message": "id required"})
		return
	}

	var meal models.Meal
	if err := global.Db.First(&meal, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 0, "message": "meal not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 0, "message": "query meal failed", "err": err.Error()})
		return
	}

	// 查询 meal_dish
	var mealDishes []models.MealDish
	if err := global.Db.Where("meal_id = ?", meal.ID).Find(&mealDishes).Error; err != nil {
		log.Printf("GetPublicMealById: meal_dish query failed: %v", err)
		mealDishes = []models.MealDish{}
	}

	// 收集 dish ids
	dishIDsMap := make(map[int]struct{})
	for _, md := range mealDishes {
		dishIDsMap[md.DishID] = struct{}{}
	}
	dishIDs := make([]int, 0, len(dishIDsMap))
	for idk := range dishIDsMap {
		dishIDs = append(dishIDs, idk)
	}

	var referencedDishes []models.Dish
	if len(dishIDs) > 0 {
		if err := global.Db.Where("id IN ?", dishIDs).Find(&referencedDishes).Error; err != nil {
			log.Printf("GetPublicMealById: referenced dishes query failed: %v", err)
			referencedDishes = []models.Dish{}
		}
	}

	dishByID := make(map[int]models.Dish)
	for _, dd := range referencedDishes {
		dishByID[dd.ID] = dd
	}

	entries := make([]map[string]interface{}, 0, len(mealDishes))
	for _, md := range mealDishes {
		d := dishByID[md.DishID]
		entries = append(entries, map[string]interface{}{
			"dishId": md.DishID,
			"name":   d.DishName,
			"price":  d.Price,
			"image":  d.ImagePath,
			"copies": md.Num,
		})
	}

	resp := map[string]interface{}{
		"id":            meal.ID,
		"name":          meal.Mealname,
		"price":         meal.Price,
		"description":   meal.Description,
		"image":         meal.ImagePath,
		"categoryId":    meal.Category,
		"status":        meal.Status,
		"setmealDishes": entries,
	}

	c.JSON(http.StatusOK, gin.H{"code": 1, "data": resp})
}
