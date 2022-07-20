package svc

import (
	"context"
	"errors"
	"fmt"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ThumbService struct {
	pkg.BaseComponent
	APIGatewayServiceClient imapigateway.ApiGatewayServiceClient
	ThumbDBServiceClient    moment.ThumbDBServiceClient
	ForumDBServiceClient    moment.ForumDBServiceClient
	CommentDBServiceClient  moment.CommentDBServiceClient
	NoticeDBServiceClient   moment.NoticeDBServiceClient
	BaseDBServiceClient     moment.BaseDBServiceClient
}

//ThumbAdd 新增点赞
// @Tags 点赞
// @Summary 新增点赞
// @Produce  json
// @Description 传参 ForumID、ForumUID、UpDown(1赞2踩) 如果是点赞的评论或回复 需要传 CommentID、CommentUID
// @Param Authorization header string true "Token"
// @Param body body moment.Thumb true "json "
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /thumbsvc/add [post]
func (ts *ThumbService) ThumbAdd(c *gin.Context) {
	var params = &moment.Thumb{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		ts.Response(c, "THUMB_ADD", err, "解析点赞参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ForumID == 0 && params.CommentID == 0 {
		ts.Response(c, "THUMB_ADD", errors.New("ForumID、CommentID参数不能为空"), "解析点赞参数", nil)
		return
	}

	// 测试要求检查帖子是否存在
	// 测试要求检查帖子是否存在
	var forum = &moment.ForumFriend{}
	if forum, err = ts.ForumDBServiceClient.Get(c.Request.Context(), &moment.ForumGetReq{ForumID: params.ForumID}); err != nil {
		ts.Response(c, "COMMENT_ADD", err, "获取帖子失败", nil)
		return
	}
	if forum.ID == 0 {
		ts.Response(c, "FORUM_NOT_FOUND", errors.New("帖子不存在"), "帖子不存在", nil)
		return
	}

	params.ForumUID = forum.CreateBy

	//点赞回复和评论
	var comment = &moment.Comment{CommentID: params.CommentID}
	if params.CommentID > 0 {
		if comment, err = ts.CommentDBServiceClient.Get(c.Request.Context(), comment); err != nil {
			ts.Response(c, "THUMB_ADD", err, "检查评论或回复出错", nil)
			return
		}
		if comment.ID == 0 {
			ts.Response(c, "THUMB_ADD", errors.New("评论不存在"), "评论不存在", nil)
			return
		}
		params.CommentUID = comment.CreateBy

	}
	//点赞帖子
	if params.CommentID == 0 {
		var userInfo *imapigateway.GetUserInfoResult
		if userInfo, err = pkg.GetUserInfoByID(c.Request.Context(), ts.APIGatewayServiceClient, params.CreateBy, params.CreateBy); err != nil {
			ts.Response(c, "THUMB_ADD", err, "查询用户信息", params)
			return
		}
		params.Creator = userInfo.Uinfo
	}

	if params, err = ts.ThumbDBServiceClient.Add(c.Request.Context(), params); err != nil {
		ts.Response(c, "THUMB_ADD", err, "新增点赞", params)
		return
	}

	// //通知IM更新 小红点
	// //查询帖子参与人
	// var friendIDs []int32
	// if friendIDs, err = pkg.GetCommonFriends(c.Request.Context(), ts.APIGatewayServiceClient, params.CreateBy, mainUID); err != nil {
	// 	ts.Log.Error("获取双向共同好友失败", zap.Int32("用户ID", params.CreateBy), zap.Int32("朋友ID", params.CreateBy), zap.Error(err))
	// }
	// friendIDs = append(friendIDs, mainUID)
	// // 查询参与帖子得共同好友
	// req := &moment.ParticipatingFriendsMsg{Friends: friendIDs, ForumID: params.f}
	// if req, err = ts.ForumDBServiceClient.ParticipatingFriends(c.Request.Context(), req); err != nil {
	// 	ts.Log.Error("获取帖子参与用户出错", zap.Int64("帖子", params.MainID), zap.Int32s("朋友ID", friendIDs), zap.Error(err))
	// }

	//通知@和提及消息
	// go func() {
	// 	var notices []*moment.Notice

	// 	for _, friend := range req.Friends {
	// 		notices = append(notices, &moment.Notice{Notifier: friend, CreateBy: params.CreateBy, RelationID: params.MainID, Type: moment.NoticeType_Thumb_Type})
	// 	}
	// 	if _, err = ts.NoticeDBServiceClient.AddBatch(c.Request.Context(), &moment.NoticeAddBatchReq{Notices: notices}); err != nil {
	// 		ts.Log.Error("新增点赞消息出错", zap.Any("notices", notices), zap.Error(err))
	// 	}
	// }()
	// 只有点赞帖子才发互动消息
	if params.CommentID == 0 && params.CreateBy != forum.CreateBy {

		//发送互动消息
		go func(params *moment.Thumb, forum *moment.ForumFriend) {
			ctx := context.TODO()
			notic := &moment.Notice{Type: moment.NoticeType_Thumb_Type, RelationID: params.ForumID, CreateBy: params.CreateBy, Notifier: forum.CreateBy, Status: 1}
			if notic, err = ts.NoticeDBServiceClient.Get(ctx, notic); err != nil {
				ts.Log.Error("查询是否已过赞出错", zap.Error(err))
				return
			}
			//已点过赞 忽略二次点赞
			if notic.ID > 0 {
				return
			}
			if _, err = ts.NoticeDBServiceClient.Add(ctx, &moment.Notice{Type: moment.NoticeType_Thumb_Type, RelationID: params.ForumID, CreateBy: params.CreateBy, Notifier: params.ForumUID, Status: 1}); err != nil {
				ts.Log.Error("新增点赞消息出错", zap.Error(err))
				return
			}

			if err = pkg.ImMomentNotify(ctx,
				ts.APIGatewayServiceClient,
				&imapigateway.MomentNotifyReq{
					MomentNotifyData: []*imapigateway.MomentNotifyData{
						{
							SourceCode:   4,
							BusinessCode: 6,
							ForumId:      forum.ForumID,
							ForumImage: func() string {
								if forum.ContentType > 1 {
									return forum.Medias[0].Thum
								}
								return ""
							}(),
							ForumType:  int64(forum.ContentType),
							ForumText:  forum.Content,
							ToId:       params.ForumUID,
							WithId:     params.CreateBy,
							IsFollowed: 2,
							MsgTime:    params.CreateAt,
							Entites: func() (entitys []*imapigateway.Entity) {
								for _, entity := range forum.Entitys {
									entitys = append(entitys, &imapigateway.Entity{Type: 7, UserId: uint32(entity.UserID), Offset: entity.UOffset, Length: entity.ULimit})
								}
								return
							}(),
						}},
				}); err != nil {
				ts.Log.Error("发送IM消息出错", zap.Error(err))
			}
		}(params, forum)
	}
	ts.Response(c, "THUMB_ADD", err, "新增点赞", params)
}

//ThumbDelete 删除点赞
// @Tags 点赞
// @Summary 删除点赞
// @Produce  json
// @Description 传参 ThumbID、ForumID、ForumUID 如果是去掉评论的点赞 需要传 CommentID、CommentUID
// @Param body body moment.Thumb true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /thumbsvc/delete [post]
func (ts *ThumbService) ThumbDelete(c *gin.Context) {
	var params = &moment.Thumb{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		ts.Response(c, "THUMB_DELETE", err, "解析取消点赞参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if (params.ForumID == 0 && params.CommentID == 0) || params.ForumUID == 0 {
		ts.Response(c, "THUMB_DELETE", errors.New("参数检测出错ForumID、CommentID、ForumUID"), "解析取消点赞", nil)
		return
	}
	fmt.Println("params>>>>>>>>>>>>>>", params.ForumID, params.ForumUID, params.CreateBy, params.CommentID, params.CommentUID)
	params, err = ts.ThumbDBServiceClient.Delete(c.Request.Context(), params)
	ts.Response(c, "THUMB_DELETE", err, "取消点赞", params)
	//删除消息
	// _, err = ts.NoticeDBServiceClient.Delete(c.Request.Context(), &moment.Notice{RelationID: params.ThumbID})
}

//ThumbPage 获取点赞
// @Tags 点赞
// @Summary 获取点赞
// @Produce  json
// @Param ThumbID  query int true "最后浏览点赞编号"
// @Param ForumID  query int true "帖子ID"
// @Param Limit  query int true "分页"
// @Param UpDown  query int true "1点赞2踩,目前只有点赞"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /thumbsvc/page [get]
func (ts *ThumbService) ThumbPage(c *gin.Context) {
	var params = &moment.ThumbPageReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		ts.Response(c, "THUMB_PAGE", err, "解析获取点赞参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ForumID == 0 {
		ts.Response(c, "THUMB_PAGE", errors.New("参数检测出错ForumID"), "解析获取点赞", nil)
		return
	}
	var forum *moment.ForumFriend
	if forum, err = ts.ForumDBServiceClient.Get(c.Request.Context(), &moment.ForumGetReq{ForumID: params.ForumID}); err != nil {
		ts.Response(c, "THUMB_PAGE", err, "查询帖子信息", nil)
		return
	}
	if forum.ID == 0 {
		ts.Response(c, "FORUM_NOT_FOUND", errors.New("帖子不存在或已删除"), "帖子不存在或已删除", nil)
		return
	}

	//检查用户是否是推荐用户
	var req = &moment.UserRecommend{UserID: forum.CreateBy}
	if req, err = ts.BaseDBServiceClient.UserRecommendGet(c.Request.Context(), req); err != nil {
		ts.Response(c, "THUMB_PAGE", err, "检查用户是否是推荐用户", nil)
		return
	}
	//推荐用户从朋友列表里面查看需要过滤好友
	//其他都过滤
	if req.ID == 0 {
		var friends []int32
		if friends, err = pkg.GetCommonFriends(c.Request.Context(), ts.APIGatewayServiceClient, pkg.GetClaims(c).UserID, forum.CreateBy); err != nil {
			ts.Response(c, "THUMB_PAGE", err, "查询共同好友", nil)
			return
		}
		params.Friends = friends
	}

	var rep = &moment.ThumbPageRep{}
	if rep, err = ts.ThumbDBServiceClient.Page(c.Request.Context(), params); err != nil {
		ts.Response(c, "THUMB_PAGE", err, "解析获取点赞", nil)
		return
	}

	//查询头像
	var users []int32
	for _, thumb := range rep.Thumbs {
		users = append(users, thumb.CreateBy)
	}
	if len(users) > 0 {
		var userInfoRep *imapigateway.GetUserInfoByIdArrResult
		if userInfoRep, err = pkg.GetUserInfoByIDArr(c.Request.Context(), ts.APIGatewayServiceClient, users, params.CreateBy); err != nil {
			ts.Response(c, "THUMB_PAGE", err, "解析用户信息", nil)
			return
		}
		rep.UserInfo = userInfoRep.Uinfo

	}

	ts.Response(c, "THUMB_PAGE", err, "获取点赞", rep)
}

//UserCount 用户获取点赞数量
// @Tags 点赞
// @Summary 获取用户被点赞数量
// @Produce  json
// @Param UserID  query int true "用户ID"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /thumbsvc/usercount [get]
func (ts *ThumbService) UserCount(c *gin.Context) {
	var params = &moment.ThumbUserCountReq{}
	var err error
	if err = c.BindQuery(params); err != nil {
		ts.Response(c, "THUMB_PAGE", err, "解析获取点赞参数", nil)
		return
	}
	// params.CreateBy := pkg.GetClaims(c).CreateBy
	if params.UserID == 0 {
		ts.Response(c, "THUMB_PAGE", errors.New("参数检测出错UserID"), "解析获取点赞", nil)
		return
	}
	var rep *moment.ThumbUserCountRep
	rep, err = ts.ThumbDBServiceClient.UserCount(c.Request.Context(), params)
	ts.Response(c, "THUMB_PAGE", err, "获取点赞", rep)
}
