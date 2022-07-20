package svc

import (
	"moments/pkg"
	"moments/pkg/protoc/moment"

	"github.com/gin-gonic/gin"
)

type TopicService struct {
	pkg.BaseComponent
	moment.TopicDBServiceClient
}

//TopicPage 获取话题
// @Tags 话题
// @Summary 获取话题
// @Produce  json
// @Param TopicName  query string false "话题名字"
// @Param TopicID  query int false "最后浏览帖子ID"
// @Param Limit  query int true "分页"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /topicsvc/page [get]
func (ts *TopicService) TopicPage(c *gin.Context) {
	var params = &moment.TopicPageReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		ts.Response(c, "TOPIC_PAGE", err, "解析获取话题参数", nil)
		return
	}
	// if params.TopicTypeID == 0 {
	// 	ts.Response(c, "TOPIC_PAGE", errors.New("TopicTypeID不能为空"), "TopicTypeID不能为空", nil)
	// 	return
	// }
	var Topics *moment.TopicPageRep
	Topics, err = ts.TopicDBServiceClient.Page(c.Request.Context(), params)
	ts.Response(c, "TOPIC_PAGE", err, "获取话题", Topics.Topics)
}

//TopicTypeAll 获取话题类型
// @Tags 话题
// @Summary 获取话题类型
// @Produce  json
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /topicsvc/type/all [get]
func (ts *TopicService) TopicTypeAll(c *gin.Context) {
	var params = &moment.TopicTypeReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		ts.Response(c, "TOPIC_PAGE", err, "解析获取话题类型参数", nil)
		return
	}

	var Topics *moment.TopicTypeRep
	Topics, err = ts.TopicDBServiceClient.Types(c.Request.Context(), params)
	ts.Response(c, "TOPIC_PAGE", err, "获取话题类型", Topics.TopicTypes)
}
