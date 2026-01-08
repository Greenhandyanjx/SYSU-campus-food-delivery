package config

import (
	"backend/global"
	"backend/models"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() {
	dsn := AppConfig.Database.DSN
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("initial dsn fail:%v", err)
	}
	sqldb, err := db.DB()
	if err != nil {
		panic("InitDb failed")
	}
	sqldb.SetMaxIdleConns(AppConfig.Database.MaxIdleConns)
	sqldb.SetMaxOpenConns(AppConfig.Database.MaxOpenConns)
	sqldb.SetConnMaxIdleTime(5 * time.Minute)
	sqldb.SetConnMaxLifetime(30 * time.Minute)
	global.Db = db

	// Try initialize Redis; ignore failure but log if needed
	if err := InitRedis(); err != nil {
		// don't fail startup, just log
		fmt.Println("warning: redis init failed:", err)
	}
}

func Initalldb() error {
	if global.Db == nil {
		log.Println("Initalldb: global.Db is nil, please call InitDB() first")
		return fmt.Errorf("global.Db is nil")
	}

	// 一个小工具函数，方便打印是哪一步出错
	migrate := func(name string, fn func(*gorm.DB) error) error {
		log.Printf(">>> AutoMigrate start: %s\n", name)
		if err := fn(global.Db); err != nil {
			log.Printf("xxx AutoMigrate failed: %s, err=%v\n", name, err)
			return fmt.Errorf("migrate %s: %w", name, err)
		}
		log.Printf("<<< AutoMigrate ok: %s\n", name)
		return nil
	}

	// 1. dish + flavor
	if err := migrate("Dish + Flavor", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Dish{}, &models.Flavor{})
	}); err != nil {
		return err
	}

	// 2. 用户/商户/骑手这些基础表
	if err := migrate("BaseUser", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.BaseUser{})
	}); err != nil {
		return err
	}

	if err := migrate("User", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.User{})
	}); err != nil {
		return err
	}

	if err := migrate("Rider", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Rider{})
	}); err != nil {
		return err
	}

	if err := migrate("Merchant", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Merchant{})
	}); err != nil {
		return err
	}

	// 商家配送配置表：保存各商家起送、配送费、配送范围
	if err := migrate("MerchantDeliveryConfig", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.MerchantDeliveryConfig{})
	}); err != nil {
		return err
	}

	// 为已有商家填充默认配送配置（若不存在）
	var merchants []models.Merchant
	if err := global.Db.Find(&merchants).Error; err == nil {
		for _, m := range merchants {
			var cnt int64
			global.Db.Model(&models.MerchantDeliveryConfig{}).Where("base_id = ?", m.BaseID).Count(&cnt)
			if cnt == 0 {
				global.Db.Create(&models.MerchantDeliveryConfig{
					BaseID:        m.BaseID,
					MinPrice:      15,
					DeliveryFee:   2,
					DeliveryRange: 2000,
				})
			}
		}
	}

	// 3. 分类
	if err := migrate("Category", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Category{})
	}); err != nil {
		return err
	}

	// Seed fixed categories (1..15) if table is empty
	var cnt int64
	global.Db.Model(&models.Category{}).Count(&cnt)
	if cnt == 0 {
		fixed := []models.Category{
			{ID: 1, Name: "招牌套餐", Type: "common", Sort: 1},
			{ID: 2, Name: "现煮粉面", Type: "common", Sort: 2},
			{ID: 3, Name: "汉堡炸鸡", Type: "common", Sort: 3},
			{ID: 4, Name: "奶茶咖啡", Type: "common", Sort: 4},
			{ID: 5, Name: "日式便当", Type: "common", Sort: 5},
			{ID: 6, Name: "烧烤烤肉", Type: "common", Sort: 6},
			{ID: 7, Name: "水果拼盘", Type: "common", Sort: 7},
			{ID: 8, Name: "精致甜品", Type: "common", Sort: 8},
			{ID: 9, Name: "家常快炒", Type: "common", Sort: 9},
			{ID: 10, Name: "粥粉面饭", Type: "common", Sort: 10},
			{ID: 11, Name: "极速配送", Type: "common", Sort: 11},
			{ID: 12, Name: "午餐推荐", Type: "common", Sort: 12},
			{ID: 13, Name: "低价满减", Type: "common", Sort: 13},
			{ID: 14, Name: "沙拉轻食", Type: "common", Sort: 14},
			{ID: 15, Name: "精致下午茶", Type: "common", Sort: 15},
		}
		for _, fc := range fixed {
			global.Db.Exec("INSERT INTO categories (id, name, type, sort) VALUES (?, ?, ?, ?)", fc.ID, fc.Name, fc.Type, fc.Sort)
		}
	}

	// 4. 其它业务表
	if err := migrate("MealDish", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.MealDish{})
	}); err != nil {
		return err
	}

	if err := migrate("Meal", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Meal{})
	}); err != nil {
		return err
	}

	if err := migrate("Order", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Order{})
	}); err != nil {
		return err
	}

	if err := migrate("ChatMessage", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.ChatMessage{})
	}); err != nil {
		return err
	}

	// 这里就是刚才 SQL 提到的那张表，大概率是它在搞事
	if err := migrate("RiderProfile", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderProfile{})
	}); err != nil {
		return err
	}

	if err := migrate("OrderDish", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.OrderDish{})
	}); err != nil {
		return err
	}

	if err := migrate("OrderMeal", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.OrderMeal{})
	}); err != nil {
		return err
	}

	// CartItem: 支持把菜品或套餐加入购物车
	if err := migrate("CartItem", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.CartItem{})
	}); err != nil {
		return err
	}

	if err := migrate("SalesStat", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.SalesStat{})
	}); err != nil {
		return err
	}

	if err := migrate("Consignee", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Consignee{})
	}); err != nil {
		return err
	}

	if err := migrate("Address", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Address{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderWallet", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderWallet{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderIncomeRecord", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderIncomeRecord{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderWithdraw", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderWithdraw{})
	}); err != nil {
		return err
	}

	if err := migrate("DeliveryRoute", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.DeliveryRoute{})
	}); err != nil {
		return err
	}

	if err := migrate("Store", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.Store{})
	}); err != nil {
		return err
	}

	// ==== Rider extended models ====
	if err := migrate("RiderIssue", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderIssue{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderReview", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderReview{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderNotification", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderNotification{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderSystemMessage", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderSystemMessage{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderWorkSettings", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderWorkSettings{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderAccountSettings", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderAccountSettings{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderNotificationSettings", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderNotificationSettings{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderVerification", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderVerification{})
	}); err != nil {
		return err
	}

	if err := migrate("RiderHeatmapPoint", func(db *gorm.DB) error {
		return db.AutoMigrate(&models.RiderHeatmapPoint{})
	}); err != nil {
		return err
	}

	return nil
}
