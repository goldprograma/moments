package svc

// import "github.com/gin-gonic/gin"

// // Routers 路由分组
// func (db *DBService) Routers(c *gin.Engine) {
// 	v := c.Group(db.Cfg.ServiceName)
// 	follow := v.Group("follow")
// 	{
// 		follow.POST("/add", db.FollowAdd)
// 		follow.GET("/get", db.FollowGet)
// 		follow.POST("/delete", db.FollowDelete)
// 		follow.GET("/page", db.FollowPage)
// 		follow.POST("/addbatch", db.FollowBatchAdd)
// 		follow.POST("/deletebatch", db.FollowBatchDel)
// 		follow.GET("/all", db.FollowAll)
// 	}
// 	forumType := v.Group("/forumType")
// 	{
// 		forumType.POST("/add", db.ForumTypeAdd)
// 		forumType.GET("/all", db.ForumTypeAll)
// 		forumType.POST("/update", db.ForumTypeUpdate)
// 	}
// 	forum := v.Group("/forum")
// 	{
// 		forum.POST("/add", db.ForumAdd)
// 		forum.GET("/all", db.ForumAll)
// 		forum.POST("/delete", db.ForumUpdate)
// 		forum.GET("/get", db.ForumGet)
// 		forum.GET("/follow", db.ForumFollowGet)
// 		forum.GET("/recommend", db.ForumRecommend)
// 		forum.GET("/followall", db.ForumFollow)

// 		forum.GET("/topic", db.ForumTopic)
// 	}
// 	thumb := v.Group("/thumb")
// 	{
// 		thumb.POST("/add", db.ThumbAdd)
// 		thumb.GET("/page", db.ThumbPage)
// 		thumb.POST("/delete", db.ThumbDelete)
// 	}

// 	comment := v.Group("/comment")
// 	{
// 		comment.POST("/add", db.CommentAdd)
// 		comment.GET("/page", db.CommentPage)
// 		comment.POST("/delete", db.CommentDelete)
// 	}

// 	notice := v.Group("/notice")
// 	{

// 		notice.GET("/unreadcount", db.NoticeUnreadCount)
// 		notice.GET("/page", db.NoticePage)
// 		notice.POST("/addbatch", db.NoticeAddBatch)

// 	}
// 	topic := v.Group("/topic")
// 	{

// 		topic.GET("/page", db.TopicPage)
// 	}
// 	topictype := v.Group("/topictype")
// 	{

// 		topictype.GET("/all", db.TopicTypeAll)
// 	}
// }
