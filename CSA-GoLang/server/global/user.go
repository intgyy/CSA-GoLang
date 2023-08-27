package global

import (
	"CSA-GoLang/server/config"
	"gorm.io/gorm"
)

var DB *gorm.DB
var ServerConfig = &config.ServerConfig{}
var NacosConfig config.NacosConfig
