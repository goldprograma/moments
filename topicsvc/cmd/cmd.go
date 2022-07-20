package cmd

import (
	"fmt"
	"moments/pkg"
	"moments/pkg/protoc/moment"
	"moments/topicsvc/svc"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	tsService := &svc.TopicService{}
	tsService.RegisterComponent(confPath)
	// http 服务
	tsService.TopicDBServiceClient = moment.NewTopicDBServiceClient(tsService.GetGRPCConn("moment", "momentdb")) //获取用户好友
	go func() {
		gin.DisableConsoleColor()
		if tsService.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()

		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", tsService.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))
		app.Use(ps.PromMiddleware())

		tsService.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(tsService.Cfg.HTTPPort))
	}()
	<-errC
}
