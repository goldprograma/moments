package svc

import (
	"moments/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (fo *ThumbService) Routers(app *gin.Engine) {
	root := app.Group(fo.Cfg.ServiceName)
	{
		root.Use(pkg.JWTAuth())

		root.POST("/add", fo.ThumbAdd)
		root.POST("/delete", fo.ThumbDelete)
		root.GET("/page", fo.ThumbPage)
		root.GET("/usercount", fo.UserCount)
		root.GET("/user", fo.ThumbPage)
	}
}
