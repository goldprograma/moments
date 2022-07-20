package cmd

import (
	"fmt"
	"gitlab.moments.im/mediasvc/svc"
	"gitlab.moments.im/pkg"
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
	ms := &svc.MediaService{}
	ms.RegisterComponent(confPath, pkg.Component_ALIOSS)
	// http 服务

	go func() {
		gin.DisableConsoleColor()
		if ms.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()
		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", ms.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))

		app.Use(ps.PromMiddleware())
		ms.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(ms.Cfg.HTTPPort))
	}()
	<-errC
}
