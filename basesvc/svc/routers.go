package svc

import (
	"moments/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (bs *BaseService) Routers(app *gin.Engine) {
	root := app.Group(bs.Cfg.ServiceName)
	root.Use(pkg.JWTAuth())

	{

		root.GET("/uploadurl", bs.GetUploadDomain)

		root.HEAD("/health")

		tag := root.Group("/tag")
		{
			tag.GET("/get", bs.TagGet)
			tag.POST("/add", bs.TagAdd)
		}

		ignore := root.Group("/ignore")
		{
			ignore.GET("/get", bs.UserIgnoreGet)
			ignore.GET("/check", bs.UserIgnoreCheck)
			ignore.POST("/add", bs.UserIgnoreAdd)
			ignore.POST("/delete", bs.UserIgnoreDelete)
		}

		user := root.Group("/user")
		{
			user.GET("/album", bs.UserAlbum)
			user.GET("/version", bs.UserVersion)
			user.POST("/homebg", bs.UserHomeBackgroundUpdate)
			user.GET("/homebg", bs.UserHomeBackgroudGet)
			user.GET("/statistics", bs.UserStatistics)
		}
	}

}
