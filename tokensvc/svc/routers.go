package svc

import (
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (bs *TokenService) Routers(app *gin.Engine) {
	// 限流中间件
	app.Use(middleware.TollboothLimiter)
	root := app.Group("/tokensvc")
	{
		root.HEAD("/health")

		root.GET("/get", bs.GetToken)
		root.Use(pkg.JWTAuth())
		root.GET("/check", bs.CheckToken)
	}
}
