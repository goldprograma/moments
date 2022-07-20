package svc

import (
	"moments/pkg/protoc/moment"

	"github.com/gin-gonic/gin"
)

type ForumYellow struct {
	ForumID int64
	UserID  int32
	Type    int
	Indexs  []int
}

//RecognizeYellow 涉黄帖子
// @Tags 帖子
// @Summary 涉黄帖子
// @Produce  json
// @Description 只传 ForumID 、Index 是指图片第几张是黄图暂时没用
// @Param body body []ForumYellow true "json "
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumservice/recognize [post]
func (fs *ForumService) RecognizeYellow(c *gin.Context) {
	var err error
	fys := make([]*ForumYellow, 0)
	if err = c.BindJSON(&fys); err != nil || len(fys) == 0 {
		fs.Response(c, "FORUM_YELLOW", err, "解析黄色帖子", nil)

		return
	}
	for _, fy := range fys {

		if _, err = fs.ForumDBServiceClient.Delete(c.Request.Context(), &moment.ForumFriend{ForumID: fy.ForumID}); err != nil {
			fs.Response(c, "FORUM_YELLOW", err, "处理黄色帖子", nil)
			return
		}
	}

	fs.Response(c, "FORUM_YELLOW", err, "处理黄色帖子", nil)
}
