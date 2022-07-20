package cmd

import (
	"fmt"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
	"moments/thumbsvc/svc"
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
	thumbService := &svc.ThumbService{}
	thumbService.RegisterComponent(confPath)
	// http 服务
	dbConn := thumbService.GetGRPCConn("moment", "momentdb")
	thumbService.APIGatewayServiceClient = imapigateway.NewApiGatewayServiceClient(thumbService.GetGRPCConn("moment", "imapigatewayserver")) //获取用户好友
	thumbService.ThumbDBServiceClient = moment.NewThumbDBServiceClient(dbConn)
	thumbService.NoticeDBServiceClient = moment.NewNoticeDBServiceClient(dbConn) //获取用户好友
	thumbService.ForumDBServiceClient = moment.NewForumDBServiceClient(dbConn)   //获取用户好友
	thumbService.BaseDBServiceClient = moment.NewBaseDBServiceClient(dbConn)
	thumbService.CommentDBServiceClient = moment.NewCommentDBServiceClient(dbConn)
	go func() {
		gin.DisableConsoleColor()
		if thumbService.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()
		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", thumbService.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))

		app.Use(ps.PromMiddleware())
		thumbService.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(thumbService.Cfg.HTTPPort))
	}()
	<-errC
}
