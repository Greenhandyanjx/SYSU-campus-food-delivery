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
	
}
