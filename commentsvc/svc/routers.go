package svc

import (
	"moments/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (fo *CommentService) Routers(app *gin.Engine) {
	root := app.Group(fo.Cfg.ServiceName)
	{
		root.Use(pkg.JWTAuth())
		root.POST("/add", fo.CommentAdd)
		root.POST("/delete", fo.CommentDelete)
		root.GET("/page", fo.CommentPage)
		root.GET("/replaypage", fo.ReplayCommentPage)
		// root.GET("/friendpage", fo.CommentPage)

	}
}
