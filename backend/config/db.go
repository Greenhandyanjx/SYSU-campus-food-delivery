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

func Initalldb(){
	if err := global.Db.AutoMigrate(&models.Dish{},&models.Flavor{}); err != nil {
		fmt.Println("dish,flavor table create fail")
		return
	}
	if err := global.Db.Table("base_users").AutoMigrate(&models.BaseUser{}); err != nil {
		fmt.Println("base_user table create fail")
		return 
	}
	if err := global.Db.Table("users").AutoMigrate(&models.User{}); err != nil {
		fmt.Println("user table create fail")
		return 
	}
    if err := global.Db.Table("riders").AutoMigrate(&models.Rider{});err != nil {
		fmt.Println("riders table create fail")
		return 
	}
	if err := global.Db.Table("merchants").AutoMigrate(&models.Merchant{});err!=nil{
		fmt.Println("merchants table create fail")
		return 
	}
	if err := global.Db.Table("categories").AutoMigrate(&models.Category{}); err != nil {
		fmt.Println("categories table create fail")
		return 
	}
	
}
