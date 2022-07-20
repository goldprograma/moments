package cmd

import (
	"fmt"
	"moments/forumsvc/svc"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
	"strconv"

	"os"
	"os/signal"
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
	foService := &svc.ForumService{}
	foService.RegisterComponent(confPath)
	dbConn := foService.GetGRPCConn("moment", "momentdb")
	// http 服务
	foService.APIGatewayServiceClient = imapigateway.NewApiGatewayServiceClient(foService.GetGRPCConn("moment", "imapigatewayserver")) //获取用户好友
	foService.BaseDBServiceClient = moment.NewBaseDBServiceClient(dbConn)                                                              //获取用户好友
	// http 服务
	foService.ForumDBServiceClient = moment.NewForumDBServiceClient(dbConn)   //获取用户好友
	foService.NoticeDBServiceClient = moment.NewNoticeDBServiceClient(dbConn) //获取用户好友
	foService.FollowDBServiceClient = moment.NewFollowDBServiceClient(dbConn)
	foService.CommentDBServiceClient = moment.NewCommentDBServiceClient(dbConn)
	foService.ThumbDBServiceClient = moment.NewThumbDBServiceClient(dbConn)
	foService.MediaDBServiceClient = moment.NewMediaDBServiceClient(dbConn)

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
