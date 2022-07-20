package cmd

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"

	"gitlab.moments.im/dbsvc/svc"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"

	"google.golang.org/grpc"
)

// Run 加载服务配置,启动服务
func Run(confPath string) {
	// var err error
	// 进程操作
	errC := make(chan error, 1)
	// 中断处理
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		errC <- fmt.Errorf("%s", <-c)
	}()
	// 注册服务
	// dbService := &svc.DBService{}
	baseService := &svc.BaseDBService{}

	baseService.RegisterComponent(confPath, pkg.Component_DB, pkg.Component_REDIS)
	mediaDBService := &svc.MediaDBService{BaseComponent: baseService.BaseComponent}
	forumService := &svc.ForumDBService{BaseComponent: baseService.BaseComponent}
	commentService := &svc.CommentDBService{BaseComponent: baseService.BaseComponent}
	thumbDBService := &svc.ThumbDBService{BaseComponent: baseService.BaseComponent}
	noticeDBService := &svc.NoticeDBService{BaseComponent: baseService.BaseComponent}
	followDBService := &svc.FollowDBService{BaseComponent: baseService.BaseComponent}
	topicDBService := &svc.TopicDBService{BaseComponent: baseService.BaseComponent}

	// commentService.BaseComponent = forumService.BaseComponent
	// commentService.DB = forumService.DB
	// commentService.Log = forumService.Log

	// dbService.RegisterComponent(confPath, pkg.Component_DB, pkg.Component_REDIS)

	//当前版本不开放推荐用户动态相关功能
	// go func() {
	// 	// if err = cache.UserStatisticsCache(baseService.Redis, baseService.DB); err != nil {
	// 	// 	errC <- err
	// 	// }
	// 	if err = cache.UserRecommendCache(baseService.Redis, baseService.DB); err != nil {
	// 		fmt.Println("---==--===== @> err ", err)
	// 		errC <- err
	// 	}
	// 	//推荐用户监听
	// }()

	//当前版本不开放推荐用户动态相关功能
	//go cache.UserRecommendCacheWatch(baseService.Redis, baseService.DB, baseService.Log)

	go func() {
		errC <- forumService.StartGRPCServer(forumService.Cfg.GRPCPort, func(s *grpc.Server) {
			moment.RegisterBaseDBServiceServer(s, baseService)
			moment.RegisterCommentDBServiceServer(s, commentService)
			moment.RegisterForumDBServiceServer(s, forumService)
			moment.RegisterThumbDBServiceServer(s, thumbDBService)
			moment.RegisterNoticeDBServiceServer(s, noticeDBService)
			moment.RegisterFollowDBServiceServer(s, followDBService)
			moment.RegisterTopicDBServiceServer(s, topicDBService)
			moment.RegisterMediaDBServiceServer(s, mediaDBService)
		})
	}()

	// ??? 无意义的http服务端
	// go func() {
	// 	errC <- http.ListenAndServe(fmt.Sprintf(":%d", baseService.Cfg.HTTPPort), http.DefaultServeMux)
	// }()

	go func() {
		http.ListenAndServe(fmt.Sprintf(":%d", baseService.Cfg.PprofAddr), nil)
	}()
	fmt.Println("terminated    1111 ", <-errC)

}
