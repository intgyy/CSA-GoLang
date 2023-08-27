package main

import (
	"CSA-GoLang/server/global"
	"CSA-GoLang/server/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func main() {
	dsn := "root:root@tcp(175.178.156.205:3306)/CSA-GoLang?charset=utf8mb4&parseTime=True&loc=Local"

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)
	var err error
	global.DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("未连接数据库")
	}
	err = global.DB.AutoMigrate(&model.User{}, &model.Goods{}, &model.Cart{})
	if err != nil {
		panic(err)
	}
}
