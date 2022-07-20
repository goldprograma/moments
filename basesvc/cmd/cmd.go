package cmd

import (
	"fmt"
	"moments/Basesvc/svc"
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
	bsService := &svc.BaseService{}
	bsService.RegisterComponent(confPath)
	dbConn := bsService.GetGRPCConn("moment", "momentdb")
	bsService.BaseDBServiceClient = moment.NewBaseDBServiceClient(dbConn)
	bsService.APIGatewayServiceClient = imapigateway.NewApiGatewayServiceClient(bsService.GetGRPCConn("moment", "imapigatewayserver"))
	bsService.ForumDBServiceClient = moment.NewForumDBServiceClient(dbConn)
	bsService.FollowDBServiceClient = moment.NewFollowDBServiceClient(dbConn)
	bsService.MediaDBServiceClient = moment.NewMediaDBServiceClient(dbConn)

	go func() {
		errC <- bsService.StartGRPCServer(bsService.Cfg.GRPCPort, func(s *grpc.Server) {
			moment.RegisterBaseServiceServer(s, bsService)
		})
	}()
	go bsService.Retry()
	go func() {
		gin.DisableConsoleColor()
		if bsService.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()
		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", bsService.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))
		app.Use(ps.PromMiddleware())

		bsService.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(bsService.Cfg.HTTPPort))
	}()
	<-errC
}
