package main

import (
	"fmt"
	"github.com/spf13/viper"
	"tsc/internal/backend_service/db"
	"tsc/internal/cfg"
)

func main() {
	// 直接指定配置文件路径（包含文件名和扩展名）
	viper.SetConfigFile("cmd/db_connect_test/config.yaml")
	// 先读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("读取配置文件失败: %v", err))
	}
	// 解析配置到结构体
	var applicationConfig cfg.ApplicationConfig
	if err := viper.Unmarshal(&applicationConfig); err != nil {
		panic(fmt.Errorf("解析配置失败: %v", err))
	}

	// MYSQL
	//cfg.GlobalServerConfig = &applicationConfig
	//_, err := db.InitDB(cfg.DB_MYSQL)
	//if err != nil {
	//	return
	//}

	cfg.GlobalServerConfig = &applicationConfig
	_, err := db.InitDB(cfg.DB_SQLITE)
	if err != nil {
		return
	}


	db.DbEngine.
}
