package global

import (
	"client/config"
	"client/proto"
)

var (
	ServerConfig  = &config.ServerConfig{}
	UserSrvClient proto.UserClient
	NacosConfig   config.NacosConfig
)
