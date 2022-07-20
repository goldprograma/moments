package svc

import (
	"moments/pkg"

	"github.com/gin-gonic/gin"
)

// Routers 路由分组
func (ts *TopicService) Routers(app *gin.Engine) {
	root := app.Group(ts.Cfg.ServiceName)
	root.Use(pkg.JWTAuth())
	{

		// root.POST("/add", ts.TopicAdd)
		// root.POST("/delete", ts.TopicDelete)
		root.GET("/page", ts.TopicPage)

	}
	topicType := root.Group("/type")
	{
		topicType.GET("/all", ts.TopicTypeAll)
	}
}
