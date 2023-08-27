package initliaze

import (
	"CSA-GoLang/server/global"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func InitDb() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local", global.ServerConfig.MysqlInfo.User, global.ServerConfig.MysqlInfo.Password, global.ServerConfig.MysqlInfo.Host, global.ServerConfig.MysqlInfo.Port, global.ServerConfig.MysqlInfo.DB)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second, // 慢 SQL 阈值
			LogLevel:      logger.Info, // Log level
			Colorful:      true,        // 禁用彩色打印
		},
	)

	global.DB, _ = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
}
