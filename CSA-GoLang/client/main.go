package main

import (
	"client/global"
	"client/initliaze"

	"fmt"
	"go.uber.org/zap"
)

func main() {
	initliaze.InitLogger()
	initliaze.InitConfig()
	initliaze.InitSrvConn()
	Router := initliaze.Routers()

	zap.S().Infof("启动服务器,端口%d", global.ServerConfig.Port)
	err := Router.Run(fmt.Sprintf(":%d", global.ServerConfig.Port))
	if err != nil {
		zap.S().Panic("启动服务器失败", err.Error)
	}
}
