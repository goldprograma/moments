package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/imapigateway"
	"gitlab.moments.im/pkg/protoc/moment"
	"gitlab.moments.im/tokensvc/svc"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
)

// Run 加载服务配置,启动服务
func Run(confPath string) {
	// 进程操作
	errC := make(chan error)
	// 中断处理
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		errC <- fmt.Errorf("%s", <-c)
	}()
	// 注册服务
	// tokenService := &svc.TokenService{}
	// tokenService.RegisterComponent(confPath)
	// grpcConn := tokenService.GetGRPCConn("moment", "imapigatewayserver")
	// tokenService.imGrpcClient = imapigateway.NewApiGatewayServiceClient(grpcConn)
	bc := pkg.BaseComponent{}
	//初始化配置文件
	bc.RegisterComponent("./config.toml")
	grpcConn := bc.GetGRPCConn("moment", "imapigatewayserver")
	//初始化im网关grpc连接
	imGrpcClient := imapigateway.NewApiGatewayServiceClient(grpcConn)
	tokenService := svc.NewTokenService(imGrpcClient, bc.Log)

	go func() {
		errC <- bc.StartGRPCServer(bc.Cfg.GRPCPort, func(s *grpc.Server) {
			moment.RegisterTokenServiceServer(s, tokenService)
		})
	}()

	go func() {
		gin.DisableConsoleColor()
		if bc.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()

		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", bc.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))

		app.Use(ps.PromMiddleware())

		tokenService.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(bc.Cfg.HTTPPort))
	}()
	<-errC
}
