package svc

import (
	"moments/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (fo *ForumService) Routers(app *gin.Engine) {
	root := app.Group(fo.Cfg.ServiceName)
	{
		root.Use(pkg.JWTAuth())

		root.POST("/add", fo.ForumAdd)
		root.POST("/delete", fo.ForumDelete)
		// root.POST("/permission", fo.ForumPermission)
		root.POST("/ignore", fo.ForumIgnore)
		root.GET("/get", fo.ForumGet)
		root.GET("/recommend", fo.ForumRecommend)
		root.GET("/follow", fo.ForumFollow)
		root.POST("/recognize", fo.RecognizeYellow)

		root.GET("/topic", fo.ForumTopic)
		root.GET("/friend", fo.ForumFriend)
		root.GET("/othermain", fo.ForumOtherMain)
		root.GET("/selfmain", fo.ForumSelfMain)

	}

}
