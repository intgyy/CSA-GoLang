package main

import (
	"CSA-GoLang/server/global"
	"CSA-GoLang/server/handler"
	"CSA-GoLang/server/initliaze"
	"CSA-GoLang/server/proto"
	"CSA-GoLang/server/utils"
	"fmt"
	"github.com/hashicorp/consul/api"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	initliaze.InitLogger()
	initliaze.InitConfig()
	initliaze.InitDb()
	Port := 50051
	Port, _ = utils.GetFreePort()
	s := grpc.NewServer()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", Port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	proto.RegisterUserServer(s, &handler.UserServer{})
	zap.S().Info("服务端开启,端口号", Port)
	//err = s.Serve(lis)
	//if err != nil {
	//	log.Fatalf("failed to serve: %v", err)
	//}
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d", global.ServerConfig.ConsulInfo.Host, global.ServerConfig.ConsulInfo.Port)
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}

	registration := new(api.AgentServiceRegistration)
	registration.Name = global.ServerConfig.ServiceName
	serviceID := fmt.Sprintf("%s", uuid.NewV4())
	registration.ID = serviceID
	registration.Port = Port
	registration.Tags = []string{"CSA", "GoLang", "Exam", "srv"}
	registration.Address = "127.0.0.1"

	err = client.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	go func() {
		err = s.Serve(lis)
		if err != nil {
			panic("failed to start grpc:" + err.Error())
		}
	}()

	//接收终止信号
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	if err = client.Agent().ServiceDeregister(serviceID); err != nil {
		zap.S().Info("注销失败")
	}
	zap.S().Info("注销成功")
}
