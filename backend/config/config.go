package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App struct {
		Name string `yaml:"name"`
		Port string    `yaml:"port"`
	} `yaml:"app"`
	Database struct {
		DSN string `yaml:"DSN"`
		MaxOpenConns int `yaml:"max_open_conns"`
		MaxIdleConns int `yaml:"max_idle_conns"`
	} `yaml:"database"`
}

var AppConfig *Config

// InitConfig reads configuration from file and unmarshals into AppConfig
func InitConfig(){
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath("config/")

	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}

	cfg := Config{}
	
	if err := viper.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	AppConfig = &cfg
	InitDB()
    // if err := global.Db.Table("orders").AutoMigrate(&models.Order{}); err != nil {
	// 	fmt.Println("orders table create fail")
	// 	panic(err)
	// }
	//只运行一次
	// err:= Initalldb()
	// if(err!=nil){
	// 	panic(err)
	// }
}