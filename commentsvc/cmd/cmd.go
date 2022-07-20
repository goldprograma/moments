package cmd

import (
	"fmt"
	"moments/commentsvc/svc"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
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
	commentService := &svc.CommentService{}
	commentService.RegisterComponent(confPath)
	// http 服务
	dbConn := commentService.GetGRPCConn("moment", "momentdb")
	commentService.CommentDBServiceClient = moment.NewCommentDBServiceClient(dbConn)                                                             //获取用户好友
	commentService.APIGatewayServiceClient = imapigateway.NewApiGatewayServiceClient(commentService.GetGRPCConn("moment", "imapigatewayserver")) //获取用户好友

	commentService.NoticeDBServiceClient = moment.NewNoticeDBServiceClient(dbConn) //获取用户好友
	commentService.ForumDBServiceClient = moment.NewForumDBServiceClient(dbConn)   //获取用户好友
	commentService.BaseDBServiceClient = moment.NewBaseDBServiceClient(dbConn)     //获取用户好友
	commentService.FollowDBServiceClient = moment.NewFollowDBServiceClient(dbConn)

	go func() {
		gin.DisableConsoleColor()
		if commentService.Cfg.RunMode == "dev" {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
		app := gin.Default()
		//普罗米修斯
		ps := pkg.NewPrometheusMonitor("moments", commentService.Cfg.ServiceName)
		app.GET("/metrics", ps.PromHandler(promhttp.Handler()))

		app.Use(ps.PromMiddleware())
		commentService.Routers(app)
		errC <- app.Run(":" + strconv.Itoa(commentService.Cfg.HTTPPort))
	}()
	<-errC
}
