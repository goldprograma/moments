package cmd

import (
	"fmt"
	"moments/followsvc/svc"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
	"os"
	"os/signal"
	"strconv"
	"syscall"

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
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
		errC <- fmt.Errorf("%s", <-c)
	}()
	// 注册服务
	foService := &svc.FollowService{}
	foService.RegisterComponent(confPath)
	dbConn := foService.GetGRPCConn("moment", "momentdb")
	foService.FollowDBServiceClient = moment.NewFollowDBServiceClient(dbConn)

	foService.NoticeDBServiceClient = moment.NewNoticeDBServiceClient(dbConn)
	foService.APIGatewayServiceClient = imapigateway.NewApiGatewayServiceClient(foService.GetGRPCConn("moment", "imapigatewayserver")) //获取用户好友

	go func() {
		errC <- foService.StartGRPCServer(foService.Cfg.GRPCPort, func(s *grpc.Server) {
			moment.RegisterFollowServiceServer(s, foService)
		})
	}()

	go func() {
		gin.DisableConsoleColor()
		if foService.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()
		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", foService.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))
		app.Use(ps.PromMiddleware())
		foService.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(foService.Cfg.HTTPPort))
	}()
	<-errC
}
