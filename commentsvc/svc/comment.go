package svc

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type CommentService struct {
	pkg.BaseComponent
	APIGatewayServiceClient imapigateway.ApiGatewayServiceClient
	BaseDBServiceClient     moment.BaseDBServiceClient
	CommentDBServiceClient  moment.CommentDBServiceClient
	ForumDBServiceClient    moment.ForumDBServiceClient
	NoticeDBServiceClient   moment.NoticeDBServiceClient
	FollowDBServiceClient   moment.FollowDBServiceClient
}

//CommentAdd 新增评论
// @Tags 评论
// @Summary 新增评论
// @Produce  json
// @Description  SupID、SupUser、 ReplayID、ReplayUID、Content 必传,评论帖子ReplayID=ForumID,有媒体需要传Medias
// @Param body body moment.Comment true "json "
// @Accept  json
// @Success 200 {object} cs.ResponseMessage
// @Failure 500 {object} cs.ResponseMessage
// @Router /commentservice/add [post]
func (cs *CommentService) CommentAdd(c *gin.Context) {
	var params = &moment.Comment{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		cs.Response(c, "COMMENT_ADD", err, "解析评论参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ReplayID == 0 || params.ReplayUID == 0 || params.Content == "" {
		cs.Response(c, "COMMENT_ADD", errors.New("ForumID、ForumUser、ReplayID、ReplayUID、Content为空"), "解析评论参数", nil)
		return
	}
	params.ContentType = func() int32 {
		if len(params.Medias) == 0 {
			return 1
		} else if len(params.Medias) == 1 && params.Medias[0].Duration > 0 {
			return 3
		}
		return 2
	}()

	// 测试要求检查帖子是否存在
	var forum = &moment.Forum{}
	if forum, err = cs.ForumDBServiceClient.Get(c.Request.Context(), &moment.ForumGetReq{ForumID: params.ForumID}); err != nil {
		cs.Response(c, "COMMENT_ADD", err, "获取帖子失败", nil)
		return
	}
	if forum.ID == 0 || forum.Status == 0 {
		cs.Response(c, "COMMENT_ADD", errors.New("帖子不存在"), "帖子不存在", nil)
		return
	}

	//回复和评论
	var comment = &moment.Comment{CommentID: params.SupID, Status: 1}
	if params.SupID > 0 {
		if comment, err = cs.CommentDBServiceClient.Get(c.Request.Context(), comment); err != nil {
			cs.Response(c, "THUMB_ADD", err, "检查评论或回复出错", nil)
			return
		}
		if comment.ID == 0 {
			cs.Response(c, "THUMB_ADD", errors.New("评论不存在"), "评论不存在", nil)
			return
		}
	}

	params.ForumID, params.ForumUser = forum.ForumID, forum.CreateBy
	//如果是系统推荐用户，并且是在评论有@则处理高亮
	var atList = make([]int32, 0, len(params.Entitys))

	var req = &moment.UserRecommend{UserID: params.ForumUser}
	if req, err = cs.BaseDBServiceClient.UserRecommendGet(c.Request.Context(), req); err != nil {
		cs.Response(c, "COMMENT_ADD", err, "检查用户是否是推荐用户", nil)
		return
	}
	//排除自己评论自己
	if params.CreateBy != params.ForumUser {
		//检查用户等级是否达标，只有达标的才能评论
		var leve *imapigateway.ReplyMemberLevelCache
		if leve, err = cs.APIGatewayServiceClient.FindMemberLevelCache(c.Request.Context(), &imapigateway.ReqMemberLevelCache{ImUserId: params.CreateBy}); err != nil {
			cs.Response(c, "COMMENT_ADD", err, "查询评论用户等级出错", nil)
			return
		}
		//
		if leve.Status != 1 {
			cs.Response(c, "COMMENT_ADD", errors.New(leve.Msg), "查询评论用户VipLevelCode出错", nil)
			return
		}
		// var vipcode int64
		// if vipcode, err = strconv.ParseInt(leve.VipLevelCode, 0, 32); err != nil {
		// 	cs.Response(c, "COMMENT_ADD", err, "VipLevelCode等级解析出错", nil)
		// 	return
		// }
		fmt.Println(">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>", leve.VipLevelCode, req.LimitVIP)

		if leve.VipLevelCode < req.LimitVIP {
			cs.Response(c, "COMMENT_PERMISSIONDENIED", errors.New("评论失败"), fmt.Sprintf("该内容来至推荐,需要达到VIP%d才可以评论", req.LimitVIP), nil)
			return
		}
	}

	if params.SupID == 0 && req.ID > 0 {

		//处理@
		realHighLights := pkg.GetAtHightLight(params.Content, params.Entitys)
		for _, entity := range realHighLights {
			atList = append(atList, entity.UserID)
		}
		params.Entitys = realHighLights

		highLightJSON, _ := json.Marshal(realHighLights)
		params.AtEntity = string(highLightJSON)
		cs.Log.Debug("@好友列表", zap.Any("好友列表", atList))

	}

	var userInfoArr *imapigateway.GetUserInfoByIdArrResult
	if userInfoArr, err = pkg.GetUserInfoByIDArr(c.Request.Context(), cs.APIGatewayServiceClient, []int32{params.CreateBy, params.ReplayUID}, forum.CreateBy); err != nil {
		cs.Response(c, "FORUM_GET", err, "查询用户信息", params)
		return
	}
	var userMapp = make(map[int32]*imapigateway.UserInfo, 2)

	for i, info := range userInfoArr.Uinfo {
		userMapp[info.UserId] = userInfoArr.Uinfo[i]
	}

	params.Creator = userMapp[params.CreateBy]
	params.ReplayUser = userMapp[params.ReplayUID]

	if params, err = cs.CommentDBServiceClient.Add(c.Request.Context(), params); err != nil {
		cs.Response(c, "COMMENT_ADD", err, "评论", nil)
		return
	}
	//通知IM更新 小红点
	//查询帖子参与人
	// var friendIDs []int32
	// if friendIDs, err = pkg.GetCommonFriends(c.Request.Context(), cs.APIGatewayService, params.CreateBy, params.ReplayUID); err != nil {
	// 	cs.Log.Error("获取双向共同好友失败", zap.Int32("用户ID", params.CreateBy), zap.Int32("朋友ID", params.ReplayUID), zap.Error(err))
	// 	return
	// }

	// friendIDs = append(friendIDs, params.ReplayUID)
	// // 查询参与帖子得共同好友
	// req := &moment.ParticipatingFriendsMsg{Friends: friendIDs, ForumID: params.ForumID}
	// if req, err = cs.ForumDBServiceClient.ParticipatingFriends(c.Request.Context(), req); err != nil {
	// 	cs.Log.Error("获取帖子参与用户出错", zap.Int64("帖子", params.ForumID), zap.Int32s("朋友ID", friendIDs), zap.Error(err))
	// 	cs.Response(c, "COMMENT_ADD", err, "获取帖子参与用户出错", params)
	// 	return
	// }
	//通知IM更新 小红点
	// go pkg.UpdateMoments(context.TODO(), cs.APIGatewayService, req.Friends, params.CreateBy)
	//自己回复自己的帖子或者 自己回复自己的回复和评论

	//不是自己回复自己

	//发送互动消息
	go func(params *moment.Comment, commentUser int32) {
		var msg []*imapigateway.MomentNotifyData

		switch params.ContentType {
		case 2:
			for range params.Medias {
				params.Content += "[图片]"
			}
		case 3:
			params.Content += "[视频]"
		}
		fmt.Println(params.CommentID, params.CreateBy, forum.CreateBy, params.SupID, params.CreateBy, commentUser)

		if (params.SupID == 0 && params.CreateBy != forum.CreateBy) || (params.SupID > 0 && params.CreateBy != commentUser) {
			//发送回复
			msg = append(msg, &imapigateway.MomentNotifyData{
				SourceCode: 4,
				BusinessCode: func() int32 {
					if params.SupID > 0 {
						return 5 //回复
					}
					return 4 // 评论
				}(),
				ForumId: params.ForumID,
				ForumImage: func() string {
					if forum.ContentType > 1 {
						return forum.Medias[0].Thum
					}
					return ""
				}(),
				ForumType:  int64(forum.ContentType),
				ForumText:  forum.Content,
				ToId:       params.ReplayUID,
				WithId:     params.CreateBy,
				IsFollowed: 2,
				MsgContent: params.Content,
				MsgTime:    params.CreateAt,
				MsgId:      params.CommentID,
			})
		}
		//处理@
		for _, user := range atList {
			if params.CreateBy != user && params.ReplayUID != user { //自己@自己不发消息 评论和@同一个人只发评论不发@
				msg = append(msg, &imapigateway.MomentNotifyData{
					SourceCode:   4,
					BusinessCode: 7, //@
					ForumId:      params.ForumID,
					ForumImage: func() string {
						if params.ContentType > 1 {
							return params.Medias[0].Thum
						}
						return ""
					}(),
					ForumType:  int64(forum.ContentType),
					ForumText:  forum.Content,
					ToId:       user,
					WithId:     params.CreateBy,
					IsFollowed: 2,
					MsgId:      params.CommentID,
					MsgContent: params.Content,
					MsgTime:    params.CreateAt,
					Entites: func() (entitys []*imapigateway.Entity) {
						for _, entity := range params.Entitys {
							entitys = append(entitys, &imapigateway.Entity{Type: 7, UserId: uint32(entity.UserID), Offset: entity.UOffset, Length: entity.ULimit})
						}
						return
					}(),
				})
			}
		}

		if len(msg) > 0 {
			if err = pkg.ImMomentNotify(context.TODO(),
				cs.APIGatewayServiceClient, &imapigateway.MomentNotifyReq{MomentNotifyData: msg}); err != nil {
				cs.Log.Error("发送IM消息出错", zap.Error(err))
			}
		}
	}(params, comment.CreateBy)

	cs.Response(c, "COMMENT_ADD", err, "新增评论", params)
}

//CommentDelete 删除评论
// @Tags 评论
// @Summary 删除评论
// @Produce  json
// @Param body body moment.Comment true "json "
// @Description  CommentID、ForumID、ForumUser 必传
// @Accept  json
// @Success 200 {object} cs.ResponseMessage
// @Failure 500 {object} cs.ResponseMessage
// @Router /commentservice/delete [post]
func (cs *CommentService) CommentDelete(c *gin.Context) {
	var params = &moment.Comment{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		cs.Response(c, "COMMENT_DELETE", err, "解析删除评论参数", nil)
		return
	}
	var userID = pkg.GetClaims(c).UserID

	if params.CommentID == 0 {
		cs.Response(c, "COMMENT_DELETE", errors.New("参数检测出错CommentID"), "解析删除评论", nil)
		return
	}
	var comment = &moment.Comment{}
	//检查评论是否存在
	if comment, err = cs.CommentDBServiceClient.Get(c.Request.Context(), &moment.Comment{CommentID: params.CommentID, Status: 1}); err != nil {
		cs.Response(c, "COMMENT_DELETE", err, "查询评论出错", nil)
		return
	}
	if comment.ID == 0 {
		cs.Response(c, "COMMENT_DELETE", errors.New("评论不存在"), "查询评论出错", nil)
		return
	}
	if comment.CreateBy != userID && params.ForumUser != comment.ForumUser {
		cs.Response(c, "COMMENT_DELETE", errors.New("无权限删除该评论"), "无权限删除该评论", nil)
		return
	}

	params.ForumID = comment.CommentID
	params.ForumUser = comment.CreateBy
	params.CreateBy = userID

	params, err = cs.CommentDBServiceClient.Delete(c.Request.Context(), params)
	cs.Response(c, "COMMENT_DELETE", err, "删除评论", params)
	// //删除消息
	// if _, err = cs.NoticeDBServiceClient.Delete(c.Request.Context(), &moment.Notice{RelationID: params.CommentID}); err != nil {
	// 	cs.Log.Error("删除未读消息失败", zap.Error(err))
	// }
}

//CommentPage 获取评论
// @Tags 评论
// @Summary 获取评论
// @Produce  json
// @Param ForumID query int false "帖子ID"
// @Param ForumUser query int false "帖子创建用户此字段必传"
// @Param CommentID  query int false "最后浏览帖子ID"
// @Param Limit  query int true "分页"
// @Param ReplayLimit  query int true "回复条数"
// @Accept  json
// @Success 200 {object} cs.ResponseMessage
// @Failure 500 {object} cs.ResponseMessage
// @Router /commentservice/page [get]
func (cs *CommentService) CommentPage(c *gin.Context) {
	var params = &moment.CommentPageReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		cs.Response(c, "COMMENT_PAGE", err, "解析获取评论参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ForumID == 0 {
		cs.Response(c, "COMMENT_PAGE", errors.New("参数检测出错ForumID"), "解析获取评论", nil)
		return
	}
	var forum *moment.Forum
	if forum, err = cs.ForumDBServiceClient.Get(c.Request.Context(), &moment.ForumGetReq{ForumID: params.ForumID}); err != nil {
		cs.Response(c, "COMMENT_PAGE", err, "查询帖子信息", nil)
		return
	}
	if forum.Status != 1 {
		cs.Response(c, "COMMENT_PAGE", errors.New("帖子不存在或已删除"), "帖子不存在或已删除", nil)
		return
	}

	//检查用户是否是推荐用户
	var req = &moment.UserRecommend{UserID: forum.CreateBy}
	if req, err = cs.BaseDBServiceClient.UserRecommendGet(c.Request.Context(), req); err != nil {
		cs.Response(c, "COMMENT_PAGE", err, "检查用户是否是推荐用户", nil)
		return
	}
	//推荐用户从朋友列表里面查看需要过滤好友
	//其他都过滤
	if req.ID == 0 {
		var friends []int32
		if friends, err = pkg.GetCommonFriends(c.Request.Context(), cs.APIGatewayServiceClient, pkg.GetClaims(c).UserID, forum.CreateBy); err != nil {
			cs.Response(c, "COMMENT_PAGE", err, "查询共同好友", nil)
			return
		}
		params.Friends = append(params.Friends, forum.CreateBy, params.CreateBy)
		params.Friends = append(params.Friends, friends...)

	}

	//不是查看自己的帖子评论
	// if forum.CreateBy != pkg.GetClaims(c).UserID {

	// 	switch forum.Permission {
	// 	case 1: //公开
	// 	case 2: //私密
	// 		cs.Response(c, "COMMENT_PAGE", errors.New("无权限"), "无权限", nil)
	// 		return
	// 	case 5: //指定好友可见
	// 		if forum.Permission == 5 {
	// 			if err = json.Unmarshal([]byte(forum.PermissionUser), &params.Friends); err != nil {
	// 				cs.Response(c, "COMMENT_PAGE", err, "查询指定可见好友", nil)
	// 				return
	// 			}
	// 		}
	// 	case 3, 6: //朋友，指定人不可见
	// 		var friends []int32
	// 		if friends, err = pkg.GetCommonFriends(c.Request.Context(), cs.APIGatewayServiceClient, pkg.GetClaims(c).UserID, forum.CreateBy); err != nil {
	// 			cs.Response(c, "COMMENT_PAGE", err, "查询共同好友", nil)
	// 			return
	// 		}

	// 		if forum.Permission == 6 {
	// 			var permissionUser []int32
	// 			if err = json.Unmarshal([]byte(forum.PermissionUser), &permissionUser); err != nil {
	// 				cs.Response(c, "COMMENT_PAGE", err, "查询指定不可见好友", nil)
	// 				return
	// 			}
	// 			var permissionUserMapper map[int32]struct{}
	// 			for i := range permissionUser {
	// 				permissionUserMapper[permissionUser[i]] = struct{}{}
	// 			}
	// 			for _, friend := range friends {
	// 				if _, ok := permissionUserMapper[friend]; !ok {
	// 					params.Friends = append(params.Friends, friend)
	// 				}
	// 			}
	// 		}
	// 	case 4: //粉丝
	// 		var fans *moment.FansAllIDRep
	// 		if fans, err = cs.FollowDBServiceClient.FansAllID(c.Request.Context(), &moment.Follow{FollowUID: forum.CreateBy}); err != nil {
	// 			cs.Response(c, "COMMENT_PAGE", err, "查询共同好友", nil)
	// 			return
	// 		}
	// 		params.Friends = fans.Fans

	// 	}

	// }

	var comments = &moment.CommentPageRep{}
	if comments, err = cs.CommentDBServiceClient.PageOrderByThumbup(c.Request.Context(), params); err != nil {
		cs.Response(c, "COMMENT_PAGE", err, "获取评论", comments)
		return
	}
	var users []int32
	if params.ReplayLimit > 0 { //查询子回复
		for i, comment := range comments.Comments {
			var replayPageResp *moment.ReplayPageResp
			users = append(users, comment.CreateBy, comment.ReplayUID)
			if replayPageResp, err = cs.CommentDBServiceClient.ReplayPage(c.Request.Context(), &moment.ReplayPageReq{CreateBy: comment.CreateBy, ForumID: comment.ForumID, SupID: comment.CommentID, Limit: params.ReplayLimit, Friends: params.Friends}); err != nil {
				cs.Response(c, "COMMENT_PAGE", err, "查询回复出错", nil)
				return
			}
			for _, comment := range replayPageResp.Comments {
				users = append(users, comment.CreateBy, comment.ReplayUID)
			}
			comments.Comments[i].SubComment = replayPageResp.Comments
		}
	}
	//查询头像
	var userInfoRep *imapigateway.GetUserInfoByIdArrResult
	if userInfoRep, err = pkg.GetUserInfoByIDArr(c.Request.Context(), cs.APIGatewayServiceClient, users, params.CreateBy); err != nil {
		cs.Response(c, "COMMENT_PAGE", err, "查询用户信息", nil)
		return
	}
	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i, userInfo := range userInfoRep.Uinfo {
		userInfoMapper[userInfo.UserId] = userInfoRep.Uinfo[i]
	}
	for i := range comments.Comments {
		comments.Comments[i].Creator = userInfoMapper[comments.Comments[i].CreateBy]
		comments.Comments[i].ReplayUser = userInfoMapper[comments.Comments[i].ReplayUID]
		//处理@
		if comments.Comments[i].AtEntity != "" {
			json.Unmarshal([]byte(comments.Comments[i].AtEntity), &comments.Comments[i].Entitys)
		}

		for j := range comments.Comments[i].SubComment {
			comments.Comments[i].SubComment[j].Creator = userInfoMapper[comments.Comments[i].SubComment[j].CreateBy]
			comments.Comments[i].SubComment[j].ReplayUser = userInfoMapper[comments.Comments[i].SubComment[j].ReplayUID]
		}
	}

	cs.Response(c, "COMMENT_PAGE", err, "获取评论", comments)
}

//ReplayCommentPage 获取回复
// @Tags 评论
// @Summary 获取回复
// @Produce  json
// @Param ForumID query int false "帖子ID"
// @Param CommentID  query int false "最后浏览帖子ID"
// @Param Limit  query int true "分页"
// @Param SupID  query int true "评论ID"
// @Param SupUID  query int true "一级评论人ID"
// @Accept  json
// @Success 200 {object} cs.ResponseMessage
// @Failure 500 {object} cs.ResponseMessage
// @Router /commentservice/replaypage [get]
func (cs *CommentService) ReplayCommentPage(c *gin.Context) {
	var params = &moment.ReplayPageReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		cs.Response(c, "COMMENT_REPLAY_PAGE", err, "解析获取评论参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.SupID == 0 || params.SupUID == 0 {
		cs.Response(c, "COMMENT_REPLAY_PAGE", errors.New("参数检测出错SupID、SupUID"), "解析获取评论", nil)
		return
	}
	var forum *moment.Forum
	if forum, err = cs.ForumDBServiceClient.Get(c.Request.Context(), &moment.ForumGetReq{ForumID: params.ForumID}); err != nil {
		cs.Response(c, "COMMENT_REPLAY_PAGE", err, "查询帖子信息", nil)
		return
	}
	if forum.Status != 1 {
		cs.Response(c, "COMMENT_REPLAY_PAGE", errors.New("帖子不存在或已删除"), "帖子不存在或已删除", nil)
		return
	}

	//检查用户是否是推荐用户
	var req = &moment.UserRecommend{UserID: forum.CreateBy}
	if req, err = cs.BaseDBServiceClient.UserRecommendGet(c.Request.Context(), req); err != nil {
		cs.Response(c, "COMMENT_REPLAY_PAGE", err, "检查用户是否是推荐用户", nil)
		return
	}
	//推荐用户从朋友列表里面查看需要过滤好友
	//其他都过滤
	if req.ID == 0 {
		var friends []int32
		if friends, err = pkg.GetCommonFriends(c.Request.Context(), cs.APIGatewayServiceClient, pkg.GetClaims(c).UserID, forum.CreateBy); err != nil {
			cs.Response(c, "COMMENT_REPLAY_PAGE", err, "查询共同好友", nil)
			return
		}
		friends = append(friends, pkg.GetClaims(c).UserID, forum.CreateBy)
		params.Friends = friends
	}

	var comments = &moment.ReplayPageResp{}
	if comments, err = cs.CommentDBServiceClient.ReplayPage(c.Request.Context(), params); err != nil {
		cs.Response(c, "COMMENT_REPLAY_PAGE", err, "获取评论", comments)
		return
	}

	//查询头像
	var users []int32
	for _, comment := range comments.Comments {
		users = append(users, comment.CreateBy, comment.ReplayUID)
	}
	if len(users) > 0 {
		var userInfoRep *imapigateway.GetUserInfoByIdArrResult
		if userInfoRep, err = pkg.GetUserInfoByIDArr(c.Request.Context(), cs.APIGatewayServiceClient, users, params.CreateBy); err != nil {
			cs.Response(c, "COMMENT_REPLAY_PAGE", err, "解析用户信息", nil)
			return
		}

		var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
		for i, userInfo := range userInfoRep.Uinfo {
			userInfoMapper[userInfo.UserId] = userInfoRep.Uinfo[i]
		}
		for i := range comments.Comments {
			comments.Comments[i].Creator = userInfoMapper[comments.Comments[i].CreateBy]
			comments.Comments[i].ReplayUser = userInfoMapper[comments.Comments[i].ReplayUID]
		}
	}

	cs.Response(c, "COMMENT_REPLAY_PAGE", err, "获取评论", comments)
}
