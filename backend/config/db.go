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
}

func Initalldb() error {
	if err := global.Db.AutoMigrate(&models.Dish{}, &models.Flavor{}); err != nil {
		fmt.Println("dish,flavor table create fail")
		return err
	}
	if err := global.Db.Table("base_users").AutoMigrate(&models.BaseUser{}); err != nil {
		fmt.Println("base_user table create fail")
		return err
	}
	if err := global.Db.Table("users").AutoMigrate(&models.User{}); err != nil {
		fmt.Println("user table create fail")
		return err
	}
	if err := global.Db.Table("riders").AutoMigrate(&models.Rider{}); err != nil {
		fmt.Println("riders table create fail")
		return err
	}
	if err := global.Db.Table("merchants").AutoMigrate(&models.Merchant{}); err != nil {
		fmt.Println("merchants table create fail")
		return err
	}
	if err := global.Db.Table("categories").AutoMigrate(&models.Category{}); err != nil {
		fmt.Println("categories table create fail")
		return err
	}

	// Seed fixed categories (1..15) if table is empty
	var cnt int64
	global.Db.Table("categories").Count(&cnt)
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
			// use Create with explicit ID
			global.Db.Exec("INSERT INTO categories (id, name, type, sort) VALUES (?, ?, ?, ?)", fc.ID, fc.Name, fc.Type, fc.Sort)
		}
	}
	if err := global.Db.Table("meal_dishes").AutoMigrate(&models.MealDish{}); err != nil {
		fmt.Println("mealdish table create fail")
		return err
	}
	// 创建 meal
	if err := global.Db.Table("meals").AutoMigrate(&models.Meal{}); err != nil {
		fmt.Println("mealdish table create fail")
		return err
	}
	if err := global.Db.Table("orders").AutoMigrate(&models.Order{}); err != nil {
		fmt.Println("orders table create fail")
		return err
	}
	// 聊天消息表
	if err := global.Db.Table("chat_messages").AutoMigrate(&models.ChatMessage{}); err != nil {
		fmt.Println("chat_messages table create fail")
		return err
	}
	if err := global.Db.Table("rider_profiles").AutoMigrate(&models.RiderProfile{}); err != nil {
		fmt.Println("rider_profiles table create fail")
		return err
	}
	if err := global.Db.Table("order-dish").AutoMigrate(&models.OrderDish{}); err != nil {
		fmt.Println("orderdish table create fail")
		panic(err)
	}
	if err := global.Db.Table("order-meal").AutoMigrate(&models.OrderMeal{}); err != nil {
		fmt.Println("ordermeal table create fail")
		panic(err)
	}
	if err := global.Db.Table("sales-stats").AutoMigrate(&models.SalesStat{}); err != nil {
		fmt.Println("salesstats table create fail")
		panic(err)
	}
	if err := global.Db.Table("consignees").AutoMigrate(&models.Consignee{}); err != nil {
		fmt.Println("consignees table create fail")
		panic(err)
	}
	if err := global.Db.Table("addresses").AutoMigrate(&models.Address{}); err != nil {
		fmt.Println("addressees table create fail")
		panic(err)
	}
	if err := global.Db.Table("orders").AutoMigrate(&models.Order{}); err != nil {
		fmt.Println("orders table create fail")
		panic(err)
	}
	if err := global.Db.Table("rider_wallets").AutoMigrate(&models.RiderWallet{}); err != nil {
		fmt.Println("rider_wallets table create fail")
		return err
	}

	if err := global.Db.Table("rider_income_records").AutoMigrate(&models.RiderIncomeRecord{}); err != nil {
		fmt.Println("rider_income_records table create fail")
		return err
	}

	if err := global.Db.Table("rider_withdraws").AutoMigrate(&models.RiderWithdraw{}); err != nil {
		fmt.Println("rider_withdraws table create fail")
		return err
	}

	if err := global.Db.Table("delivery_routes").AutoMigrate(&models.DeliveryRoute{}); err != nil {
		fmt.Println("delivery_routes table create fail")
		return err
	}
	return nil
}
