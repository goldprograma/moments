package svc

import (
	"gitlab.moments.im/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (ms *MediaService) Routers(app *gin.Engine) {
	root := app.Group(ms.Cfg.ServiceName)

	{
		root.GET("/download/:name", ms.DownloadAliOSS)
		root.Use(pkg.JWTAuth())
		root.POST("/upload", ms.Upload)

	}

}
