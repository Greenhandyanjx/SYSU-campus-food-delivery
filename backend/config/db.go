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
	return nil
}
