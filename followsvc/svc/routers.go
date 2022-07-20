package svc

import (
	"moments/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (fo *FollowService) Routers(app *gin.Engine) {
	service := app.Group(fo.Cfg.ServiceName)
	{
		service.Use(pkg.JWTAuth())
		service.POST("/add", fo.FollowAdd)
		// service.POST("/addbatch", fo.FollowAddBatch)
		service.POST("/delete", fo.FollowDelete)
		// service.POST("/deletebatch", fo.FollowDeleteBatch)
		service.GET("/get", fo.FollowGet)
		service.GET("/check", fo.FollowCheck)
		service.GET("/me", fo.Me)
		service.GET("/fans", fo.HTTPFans)
		service.GET("/fanscount", fo.HTTPFansCount)
		service.GET("/mecount", fo.MeCount)

	}
}
