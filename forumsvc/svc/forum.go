package svc

import (
	"context"
	"encoding/json"
	"errors"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ForumService struct {
	pkg.BaseComponent
	// UpdateServiceCli infoserver.ContactPushServiceClient
	NoticeDBServiceClient   moment.NoticeDBServiceClient
	ForumDBServiceClient    moment.ForumDBServiceClient
	BaseDBServiceClient     moment.BaseDBServiceClient
	FollowDBServiceClient   moment.FollowDBServiceClient
	CommentDBServiceClient  moment.CommentDBServiceClient
	ThumbDBServiceClient    moment.ThumbDBServiceClient
	MediaDBServiceClient    moment.MediaDBServiceClient
	APIGatewayServiceClient imapigateway.ApiGatewayServiceClient
}

//ForumAdd 新增帖子
// @Tags 帖子
// @Summary 新增帖子
// @Produce  json
// @Description  Content、如有媒体需要传Medias、如有提及需要Mention 内容为userid的json数组"[1]" 如果内容有@则需要传Entitys里面的UserID、UserName、AccessHash
// @Description  如果选择话题需要传Topic 为[{"TopicID": 10,"TopicName": "皇马俱乐部"}]
// @Description  "Permission": 权限1公开2私密3朋友4粉丝5指定部分可见6不给谁看，5和6需要传 "PermissionUser": "[UserID,]",
// @Param Authorization header string true "Token"
// @Param body body moment.ForumFriend true "json "
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/add [post]
func (fo *ForumService) ForumAdd(c *gin.Context) {
	var params = &moment.ForumFriend{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		fo.Response(c, "FORUM_ADD", err, "解析发帖参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	params.UserID = params.CreateBy
	if params.Content == "" && len(params.Medias) == 0 {
		fo.Response(c, "FORUM_ADD", errors.New("内容不能为空"), "发帖", nil)
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

	//处理@
	var atList = make([]int32, 0, len(params.Entitys))

	realHighLights := pkg.GetAtHightLight(params.Content, params.Entitys)
	params.Entitys = realHighLights

	highLightJSON, _ := json.Marshal(realHighLights)
	params.ContentEntity = string(highLightJSON)

	for _, entity := range realHighLights {
		atList = append(atList, entity.UserID)
	}
	// fo.Log.Debug("@好友列表", zap.Any("好友列表", atList))

	//提及
	var mentionUsers []int32
	if params.Mention != "" {
		if err = json.Unmarshal([]byte(params.Mention), &mentionUsers); err != nil {
			fo.Response(c, "FORUM_ADD", err, "解析提及", params)
			return
		}
	}

	//检查我是否是推荐用户
	var userRecommend = &moment.UserRecommend{UserID: params.CreateBy}
	if userRecommend, err = fo.BaseDBServiceClient.UserRecommendGet(context.TODO(), userRecommend); err != nil {
		fo.Response(c, "FORUM_ADD", err, "检查是否推荐号", params)
		return
	}
	params.IsRecommend = userRecommend.ID > 0
	//查询我的粉丝
	var fansAllIDReq *moment.FansAllIDRep
	if fansAllIDReq, err = fo.FollowDBServiceClient.FansAllID(context.TODO(), &moment.Follow{FollowUID: params.CreateBy}); err != nil {
		fo.Response(c, "FORUM_ADD", err, "查询我的所有粉丝", nil)
		return
	}

	//查询我的双向好友
	var friends []int32
	if friends, err = pkg.GetDoubleDiffusionFriend(context.TODO(), fo.APIGatewayServiceClient, params.CreateBy); err != nil {
		fo.Response(c, "FORUM_ADD", err, "查询我的双向好友", nil)
		return
	}
	//查询忽略我的和我忽略的
	var igResp *moment.UserIgnoresAllResp
	if igResp, err = fo.BaseDBServiceClient.UserIgnoresAll(context.TODO(), &moment.UserIgnore{UserID: params.CreateBy}); err != nil {
		fo.Response(c, "FORUM_ADD", err, "获取不让我看不到的人", nil)
		return
	}
	var igUsers = make(map[int32]struct{}, 0)
	for i := range igResp.UserIDs {
		igUsers[igResp.UserIDs[i]] = struct{}{}
	}

	for i, friend := range friends {
		if _, ok := igUsers[friend]; ok {
			friends[i] = 0
		}
	}

	//先插入到自己的列表里面
	if params, err = fo.ForumDBServiceClient.AddMySelf(c.Request.Context(), params); err != nil {
		fo.Response(c, "FORUM_ADD_ME", err, "发帖", params)
		return
	}

	var createUser *imapigateway.GetUserInfoResult
	if createUser, err = pkg.GetUserInfoByID(context.TODO(), fo.APIGatewayServiceClient, params.CreateBy, params.CreateBy); err != nil {
		fo.Response(c, "FORUM_ADD", err, "查询发帖人的用户信息", params)
		return
	}
	params.Creator = createUser.Uinfo

	//如果是推荐账号，同步插入到推荐
	if params.IsRecommend && params.Permission == 1 {
		if _, err = fo.ForumDBServiceClient.AddRecommend(c.Request.Context(), &moment.ForumRecommend{
			ForumID:         params.ForumID,
			ContentType:     params.ContentType,
			Type:            params.Type,
			Content:         params.Content,
			Longitude:       params.Longitude,
			Latitude:        params.Latitude,
			LocationCity:    params.LocationCity,
			LocationName:    params.LocationName,
			LocationAddress: params.LocationAddress,
			CreateAt:        params.CreateAt,
			CreateBy:        params.CreateBy,
			ContentEntity:   params.ContentEntity,
			Mention:         params.Mention,
			Topic:           params.Topic,
		}); err != nil {
			fo.Response(c, "FORUM_ADD_RECOMMEND", err, "发帖", params)
			return
		}
	}

	//其他异步插入
	go func(params *moment.ForumFriend) {

		switch params.Permission {
		case 1: //公开
			if params.Topic != nil { //公开才能
				if _, err = fo.ForumDBServiceClient.AddTopic(context.TODO(), &moment.ForumTopic{
					ForumID:         params.ForumID,
					ContentType:     params.ContentType,
					Type:            params.Type,
					Content:         params.Content,
					Longitude:       params.Longitude,
					Latitude:        params.Latitude,
					LocationCity:    params.LocationCity,
					LocationName:    params.LocationName,
					LocationAddress: params.LocationAddress,
					CreateAt:        params.CreateAt,
					CreateBy:        params.CreateBy,
					ContentEntity:   params.ContentEntity,
					Mention:         params.Mention,
					Topic:           params.Topic,
				}); err != nil {
					fo.Log.Error("插入话题帖子出错", zap.Error(err))
				}
			}

			for _, friend := range friends {
				params.UserID = friend
				if _, err = fo.ForumDBServiceClient.AddFriend(context.TODO(), params); err != nil {
					fo.Log.Error("扩散朋友帖子出错", zap.Int32("用户", friend), zap.Int64("帖子ID", params.ForumID), zap.Error(err))
				}
			}

		case 3: //朋友
			for _, friend := range friends {
				params.UserID = friend
				if _, err = fo.ForumDBServiceClient.AddFriend(context.TODO(), params); err != nil {
					fo.Log.Error("扩散朋友帖子出错", zap.Int32("用户", friend), zap.Int64("帖子ID", params.ForumID), zap.Error(err))
				}
			}
		case 4: //粉丝 既是好友又是粉丝朋友列表才能看
			var fans = make(map[int32]struct{}, 0)
			for _, fansUID := range fansAllIDReq.Fans {
				fans[fansUID] = struct{}{}
			}

			for _, friend := range friends {
				if _, ok := fans[friend]; ok {
					params.UserID = friend
					if _, err = fo.ForumDBServiceClient.AddFriend(context.TODO(), params); err != nil {
						fo.Log.Error("扩散朋友帖子出错", zap.Int32("用户", friend), zap.Int64("帖子ID", params.ForumID), zap.Error(err))
					}
				}
			}

		case 5: //指定部分人可见
			var userIDs = make([]int32, 0)
			if err = json.Unmarshal([]byte(params.PermissionUser), &userIDs); err != nil {
				fo.Log.Error("解析部分可见失败", zap.Error(err))
			}
			for _, friend := range userIDs {
				params.UserID = friend
				if _, err = fo.ForumDBServiceClient.AddFriend(context.TODO(), params); err != nil {
					fo.Log.Error("扩散部分朋友可见帖子出错", zap.Int32("用户", friend), zap.Int64("帖子ID", params.ForumID), zap.Error(err))
				}
			}
		case 6: //部分人不可见
			var userIDs = make([]int32, 0)
			if err = json.Unmarshal([]byte(params.PermissionUser), &userIDs); err != nil {
				fo.Log.Error("解析部分不可见失败", zap.Error(err))
			}
			var igUsers = make(map[int32]struct{}, 0)
			for _, userID := range userIDs {
				igUsers[userID] = struct{}{}
			}

			for _, friend := range friends {
				if _, ok := igUsers[friend]; !ok {
					params.UserID = friend
					if _, err = fo.ForumDBServiceClient.AddFriend(context.TODO(), params); err != nil {
						fo.Log.Error("扩散部分朋友不可见帖子出错", zap.Int32("用户", friend), zap.Int64("帖子ID", params.ForumID), zap.Error(err))
					}
				}
			}
		}

	}(params)
	// 处理小红点逻辑
	go func(userID int32, friends []int32, fans []int32) {

		var momentUserNews []*imapigateway.MomentState
		//检查是否推荐号，如果是推荐号 所有人都需要发小红点
		//获取好友 通知好友
		//获取关注 通知关注人

		//判断帖子权限
		switch params.Permission {
		case 1: //所有人

			//非推荐号//通知朋友粉丝
			if len(friends) > 0 {
				momentUserNews = append(momentUserNews, &imapigateway.MomentState{UserId: friends, Type: 2})
			}
			if len(fans) > 0 {
				momentUserNews = append(momentUserNews, &imapigateway.MomentState{UserId: fans, Type: 3})
			}
		case 3: //朋友
			if len(friends) > 0 {
				mentionUsers = append(mentionUsers, friends...)
			}
		case 4: //粉丝
			if len(fans) > 0 {
				momentUserNews = append(momentUserNews, &imapigateway.MomentState{UserId: fans, Type: 3})
			}
		case 5: //指定谁可见
			var permissionUser []int32
			if err = json.Unmarshal([]byte(params.PermissionUser), &permissionUser); err != nil {
				fo.Response(c, "FORUM_ADD", err, "解析指定观看用户格式错误", nil)
				return
			}
			momentUserNews = append(momentUserNews, &imapigateway.MomentState{UserId: permissionUser, Type: 2})
		case 6: //指定谁不可见
			var friends []int32
			if friends, err = pkg.GetDoubleDiffusionFriend(context.TODO(), fo.APIGatewayServiceClient, params.CreateBy); err != nil {
				fo.Log.Error("查询我的双向好友失败", zap.Error(err))
				return
			}
			var permissionUsers []int32
			if err = json.Unmarshal([]byte(params.PermissionUser), &permissionUsers); err != nil {
				fo.Response(c, "FORUM_ADD", err, "解析指定不可观看用户格式错误", nil)
				return
			}
			var friendMap = make(map[int32]int)
			for i, friend := range friends {
				friendMap[friend] = i
			}
			j := 0
			for _, permissionUser := range permissionUsers {
				if _, ok := friendMap[permissionUser]; !ok {
					friends[j] = permissionUser
					j++
				}
			}

			momentUserNews = append(momentUserNews, &imapigateway.MomentState{UserId: friends[:j], Type: 2})

		}
		//通知IM更新 小红点
		if err = pkg.UpdateMoments(context.TODO(), fo.APIGatewayServiceClient, momentUserNews, params.CreateBy); err != nil {
			fo.Log.Error("小红点消息出错", zap.Error(err))
			return
		}
		//更新版本号

		if _, err = fo.BaseDBServiceClient.UserVersionUpdate(context.TODO(), &moment.UserVersionUpdateReq{UserID: userID, IsRecommedUser: userRecommend.ID > 0, Fans: fans, Friend: friends, Version: params.ForumID}); err != nil {
			fo.Log.Error("更新用户版本号出错", zap.Error(err))
			return
		}

	}(params.CreateBy, friends, fansAllIDReq.Fans)

	//通知@和提及消息
	go func(params *moment.ForumFriend) {
		var msg []*imapigateway.MomentNotifyData

		// var notices = make([]*moment.Notice, 0)
		for _, atUser := range atList {
			if params.CreateBy != atUser {
				msg = append(msg, &imapigateway.MomentNotifyData{
					SourceCode:   4,
					BusinessCode: 7,
					ForumId:      params.ForumID,
					ForumImage: func() string {
						if params.ContentType > 1 {
							return params.Medias[0].Thum
						}
						return ""
					}(),
					ForumType: int64(params.ContentType),
					ForumText: params.Content,
					ToId:      atUser,
					WithId:    params.CreateBy,
					MsgTime:   params.CreateAt,
					Entites: func() (entitys []*imapigateway.Entity) {
						for _, entity := range params.Entitys {
							entitys = append(entitys, &imapigateway.Entity{Type: 7, UserId: uint32(entity.UserID), Offset: entity.UOffset, Length: entity.ULimit})
						}
						return
					}(),
				})
			}
		}

		for _, mtUser := range mentionUsers {
			msg = append(msg, &imapigateway.MomentNotifyData{
				SourceCode:   4,
				BusinessCode: 9,
				ForumId:      params.ForumID,
				ForumImage: func() string {
					if params.ContentType > 1 {
						return params.Medias[0].Thum
					}
					return ""
				}(),
				ForumType: int64(params.ContentType),
				ForumText: params.Content,
				ToId:      mtUser,
				WithId:    params.CreateBy,
				MsgTime:   params.CreateAt,
				Entites: func() (entitys []*imapigateway.Entity) {
					for _, entity := range params.Entitys {
						entitys = append(entitys, &imapigateway.Entity{Type: 7, UserId: uint32(entity.UserID), Offset: entity.UOffset, Length: entity.ULimit})
					}
					return
				}(),
			})
		}
		if len(msg) > 0 {
			if err = pkg.ImMomentNotify(context.TODO(),
				fo.APIGatewayServiceClient, &imapigateway.MomentNotifyReq{MomentNotifyData: msg}); err != nil {
				fo.Log.Error("发送IM消息出错", zap.Error(err))
			}
		}

	}(params)

	fo.Response(c, "FORUM_ADD", err, "新增帖子", params)
}

//ForumPermission 删除帖子
// @Tags 帖子
// @Summary 设置帖子权限 1公开 2私密
// @Produce  json
// @Description 只传 ForumID、Permission
// @Param body body moment.ForumFriend true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/permission [post]
// func (fo *ForumService) ForumPermission(c *gin.Context) {
// 	var params = &moment.ForumFriend{}
// 	var err error
// 	if err = c.BindJSON(&params); err != nil {
// 		fo.Response(c, "FORUM_PRIAVTE", err, "解析设置私密参数", nil)
// 		return
// 	}
// 	params.CreateBy = pkg.GetClaims(c).UserID
// 	if params.ForumID == 0 || params.Permission == 0 {
// 		fo.Response(c, "FORUM_PRIAVTE", errors.New("参数检测出错ForumID、Permission"), "解析设置私密", nil)
// 		return
// 	}

// 	_, err = fo.ForumDBServiceClient.Private(c.Request.Context(), params)
// 	fo.Response(c, "FORUM_PRIAVTE", err, "设置私密", params)
// }

//ForumIgnore 忽略帖子
// @Tags 帖子
// @Summary 忽略帖子
// @Produce  json
// @Description 只传 ForumID
// @Param body body moment.ForumIgnore true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/ignore [post]
func (fo *ForumService) ForumIgnore(c *gin.Context) {
	var params = &moment.ForumIgnore{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		fo.Response(c, "FORUM_IGNORE", err, "解析忽略帖子参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ForumID == 0 {
		fo.Response(c, "FORUM_IGNORE", errors.New("参数检测出错ForumID"), "解析忽略帖子", nil)
		return
	}

	_, err = fo.ForumDBServiceClient.Ignore(c.Request.Context(), params)
	fo.Response(c, "FORUM_IGNORE", err, "忽略帖子", params)
}

//ForumDelete 删除帖子
// @Tags 帖子
// @Summary 删除帖子
// @Produce  json
// @Description 只传 ForumID
// @Param body body moment.ForumFriend true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/delete [post]
func (fo *ForumService) ForumDelete(c *gin.Context) {
	var params = &moment.ForumFriend{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		fo.Response(c, "FORUM_DELETE", err, "解析删除帖子参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ForumID == 0 {
		fo.Response(c, "FORUM_DELETE", errors.New("参数检测出错ForumID"), "解析删除帖子", nil)
		return
	}

	_, err = fo.ForumDBServiceClient.Delete(c.Request.Context(), params)
	fo.Response(c, "FORUM_DELETE", err, "删除帖子", params)
}

//ForumGet 获取单个帖子
// @Tags 帖子
// @Summary 获取单个帖子
// @Produce  json
// @Param ForumID  query int true "帖子ID"
// // @Param CommentLimit  query int true "评论条数"
// // @Param CommentSort  query string false "评论排序asc、desc"
// // @Param CommentOrder  query string false "评论排序db字段comment_id"
// // @Param ReplayLimit  query int true "回复条数"
// // @Param ReplaySort  query string false "评论排序asc、desc"
// // @Param ReplayOrder  query string false "评论排序db字段comment_id"
// // @Param IsFriend  query bool false "如果是朋友tab需要传true/用于只查询朋友的回复"
// // @Param ThumbupLimit  query string false "点赞条数"
// // @Param ForumUser  query int true "发帖人ID"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/get [get]
func (fo *ForumService) ForumGet(c *gin.Context) {
	var params = &moment.ForumGetReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		fo.Response(c, "FORUM_GET", err, "解析获取帖子参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	if params.ForumID == 0 {
		fo.Response(c, "FORUM_GET", errors.New("参数检测出错ForumID"), "解析获取帖子", nil)
		return
	}
	var forum *moment.ForumFriend
	if forum, err = fo.ForumDBServiceClient.Get(c.Request.Context(), params); err != nil {
		fo.Response(c, "FORUM_GET", err, "查询帖子失败", nil)
		return
	}
	if forum.ID == 0 {
		fo.Response(c, "FORUM_NOT_FOUND", errors.New("帖子不存在或已删除"), "帖子不存在或已删除", nil)
		return
	}

	//检查帖子是不是推荐用户
	var userRecommend = &moment.UserRecommend{UserID: forum.CreateBy}
	if userRecommend, err = fo.BaseDBServiceClient.UserRecommendGet(context.TODO(), userRecommend); err != nil {
		fo.Response(c, "FORUM_GET", err, "检查是否推荐帖", params)
		return
	}
	forum.IsRecommend = userRecommend.ID > 0
	forum.LimitVIP = userRecommend.LimitVIP
	//如果不是 则过滤朋友权限
	if !forum.IsRecommend {
		if params.Friends, err = pkg.GetCommonFriends(c.Request.Context(), fo.APIGatewayServiceClient, pkg.GetClaims(c).UserID, forum.CreateBy); err != nil {
			fo.Response(c, "FORUM_GET", err, "查询共同好友失败", nil)
			return
		}
		params.Friends = append(params.Friends, forum.CreateBy, params.CreateBy)
	}
	//查点赞ThumbCheck
	var hasClickResp *moment.HasClickResp
	if hasClickResp, err = fo.ThumbDBServiceClient.HasClick(c.Request.Context(), &moment.Thumb{CreateBy: params.CreateBy, ForumID: forum.ForumID}); err != nil {
		fo.Response(c, "FORUM_GET", err, "检查是否点赞", nil)
		return
	}

	forum.HasThumb = hasClickResp.State

	var rep *imapigateway.GetUserInfoResult
	if rep, err = pkg.GetUserInfoByID(c.Request.Context(), fo.APIGatewayServiceClient, forum.CreateBy, forum.CreateBy); err != nil {
		fo.Response(c, "FORUM_GET", err, "获取发帖人信息出错", nil)
		return
	}
	forum.Creator = rep.Uinfo

	//处理@
	if forum.ContentEntity != "" {
		json.Unmarshal([]byte(forum.ContentEntity), &forum.Entitys)
	}
	fo.Response(c, "FORUM_GET", err, "获取帖子", forum)

}

type ForumReq struct {
	Limit    int64
	ForumID  int64
	CreateBy int32
}

//ForumRecommend 推荐帖子
// @Tags 帖子
// @Summary 推荐帖子
// @Produce  json
// @Accept  json
// @Param Limit  query int true "查询条数"
// @Param ForumID  query int true "上次浏览ID"
// @Param Authorization header string true "Token"
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/recommend [get]
func (fo *ForumService) ForumRecommend(c *gin.Context) {
	var req = &moment.RecommendPageReq{}
	var err error
	if err = c.BindQuery(req); err != nil {
		fo.Response(c, "FORUM_RECOMMEND", err, "解析推荐帖子参数", nil)
		return
	}
	var fs = &moment.RecommendPageRep{}
	req.UserID = pkg.GetClaims(c).UserID
	if fs, err = fo.ForumDBServiceClient.RecommendPage(c.Request.Context(), req); err != nil {
		fo.Response(c, "FORUM_RECOMMEND", err, "查询推荐帖子", nil)
		return
	}
	//发帖人个人信息
	var beGetUser = []int32{}
	for i := range fs.Forums {
		beGetUser = append(beGetUser, fs.Forums[i].CreateBy)
		//获取评论的用户信息
		for _, comment := range fs.Forums[i].Comments {
			beGetUser = append(beGetUser, comment.CreateBy, comment.ReplayUID)
		}
		//查询点赞的用户信息
		beGetUser = append(beGetUser, fs.Forums[i].ThumbUsers...)
	}

	var result *imapigateway.GetUserInfoByIdArrResult
	//查询点赞的用户信息
	if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, beGetUser, pkg.GetClaims(c).UserID); err != nil {
		fo.Response(c, "FORUM_RECOMMEND", err, "获取IM用户信息", nil)
		return
	}

	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i := range result.Uinfo {
		userInfoMapper[result.Uinfo[i].UserId] = result.Uinfo[i]
	}

	for i := range fs.Forums {
		fs.Forums[i].IsRecommend = true
		fs.Forums[i].Creator = userInfoMapper[fs.Forums[i].CreateBy]
		if fs.Forums[i].ContentEntity != "" {
			json.Unmarshal([]byte(fs.Forums[i].ContentEntity), &fs.Forums[i].Entitys)
		}
		//获取评论的用户信息
		for j := range fs.Forums[i].Comments {
			fs.Forums[i].Comments[j].Creator = userInfoMapper[fs.Forums[i].Comments[j].CreateBy]
			fs.Forums[i].Comments[j].ReplayUser = userInfoMapper[fs.Forums[i].Comments[j].ReplayUID]
			if fs.Forums[i].Comments[j].AtEntity != "" {
				json.Unmarshal([]byte(fs.Forums[i].Comments[j].AtEntity), &fs.Forums[i].Comments[j].Entitys)
			}
		}
		for k := range fs.Forums[i].ThumbUserInfos {
			fs.Forums[i].ThumbUserInfos[k] = userInfoMapper[fs.Forums[i].ThumbUsers[k]]
		}
	}

	fo.Response(c, "FORUM_RECOMMEND", err, "获取推荐帖子", fs.Forums)
}

//ForumTopic 话题帖子
// @Tags 帖子
// @Summary 话题帖子
// @Produce  json
// @Accept  json
// @Param Limit  query int true "查询条数"
// @Param ForumID  query int true "上次浏览ID"
// @Param TopicID  query int true "话题ID"
// @Param Authorization header string true "Token"
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/topic [get]
func (fo *ForumService) ForumTopic(c *gin.Context) {
	var req = &moment.ForumTopicPageReq{}
	var err error
	if err = c.BindQuery(req); err != nil {
		fo.Response(c, "FORUM_TOPIC", err, "解析话题帖子参数", nil)
		return
	}
	//话题帖子全公开
	var fs = &moment.ForumTopicPageRep{}
	if fs, err = fo.ForumDBServiceClient.TopicPage(c.Request.Context(), req); err != nil {
		fo.Response(c, "FORUM_TOPIC", err, "获取话题帖子失败", err)
		return
	}

	//发帖人个人信息
	var beGetUser = []int32{}
	for i := range fs.Forums {
		//检查发帖人是否是推荐用户
		var userRecommend = &moment.UserRecommend{UserID: fs.Forums[i].CreateBy}
		if userRecommend, err = fo.BaseDBServiceClient.UserRecommendGet(context.TODO(), userRecommend); err != nil {
			fo.Response(c, "FORUM_TOPIC", err, "检查是否推荐号", nil)
			return
		}
		fs.Forums[i].IsRecommend = userRecommend.ID > 0

		beGetUser = append(beGetUser, fs.Forums[i].CreateBy)
		//获取评论的用户信息
		for _, comment := range fs.Forums[i].Comments {
			beGetUser = append(beGetUser, comment.CreateBy, comment.ReplayUID)
		}
		//查询点赞的用户信息
		beGetUser = append(beGetUser, fs.Forums[i].ThumbUsers...)
	}

	var result *imapigateway.GetUserInfoByIdArrResult
	//查询点赞的用户信息
	if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, beGetUser, pkg.GetClaims(c).UserID); err != nil {
		fo.Response(c, "FORUM_RECOMMEND", err, "获取IM用户信息", nil)
		return
	}
	//
	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i := range result.Uinfo {
		userInfoMapper[result.Uinfo[i].UserId] = result.Uinfo[i]
	}

	for i := range fs.Forums {
		fs.Forums[i].Creator = userInfoMapper[fs.Forums[i].CreateBy]
		if fs.Forums[i].ContentEntity != "" {
			json.Unmarshal([]byte(fs.Forums[i].ContentEntity), &fs.Forums[i].Entitys)
		}
		//获取评论的用户信息
		for j := range fs.Forums[i].Comments {
			fs.Forums[i].Comments[j].Creator = userInfoMapper[fs.Forums[i].Comments[j].CreateBy]
			fs.Forums[i].Comments[j].ReplayUser = userInfoMapper[fs.Forums[i].Comments[j].ReplayUID]
			if fs.Forums[i].Comments[j].AtEntity != "" {
				json.Unmarshal([]byte(fs.Forums[i].Comments[j].AtEntity), &fs.Forums[i].Comments[j].Entitys)
			}
		}
		for k := range fs.Forums[i].ThumbUserInfos {
			fs.Forums[i].ThumbUserInfos[k] = userInfoMapper[fs.Forums[i].ThumbUsers[k]]
		}
	}

	fo.Response(c, "FORUM_TOPIC", err, "获取话题帖子", fs.Forums)
}

//ForumFollow 所有关注人帖子
// @Tags 帖子
// @Summary 所有关注人帖子
// @Produce  json
// @Accept  json
// @Param Limit  query int true "查询条数"
// @Param ForumID  query int true "上次浏览ID"
// @Param Authorization header string true "Token"
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/follow [get]
func (fo *ForumService) ForumFollow(c *gin.Context) {
	var req = &moment.FollowFourmPageReq{}
	var err error
	if err = c.BindQuery(req); err != nil {
		fo.Response(c, "FORUM_FOLLOW", err, "解析关注帖子参数", nil)
		return
	}

	req.UserID = pkg.GetClaims(c).UserID

	// 查询我关注的人
	var followResp *moment.FollowAllOrderByCreateAtResp
	if followResp, err = fo.FollowDBServiceClient.FollowAllOrderByCreateAt(c.Request.Context(), &moment.FollowAllOrderByCreateAtReq{UserID: req.UserID, Limit: req.Limit, CreateAt: req.CreateAt}); err != nil {
		fo.Response(c, "FORUM_FOLLOW", err, "获取我关注的人", nil)
		return
	}

	var result []*moment.ForumFriend

	var checkFriendResp *imapigateway.CheckIsMuteContactResult

	for _, follow := range followResp.Follows {
		if req.Limit <= 0 {
			break
		}
		var friend = &moment.ForumFriend{CreateBy: follow.FollowUID}
		if friend, err = fo.ForumDBServiceClient.GetNewestFromCache(c.Request.Context(), friend); err != nil {
			fo.Response(c, "FORUM_FOLLOW", err, "查询关注用户最新帖子", nil)
			return
		}
		//判断是否设置了不看

		switch friend.Permission {
		case 1, 3: //公开粉丝
			result = append(result, friend)
			req.Limit--
		case 4: //仅朋友可见
			//检查我与发帖人是否是好友
			if checkFriendResp, err = fo.APIGatewayServiceClient.CheckIsMuteContact(c.Request.Context(), &imapigateway.CheckIsMuteContactReq{SelfId: req.UserID, UserId: friend.CreateBy}); err != nil {
				fo.Response(c, "FORUM_FOLLOW", err, "检查是否是双向好友", nil)
				return
			}
			if checkFriendResp.Result {
				result = append(result, friend)
				req.Limit--
			}
		case 5: //部分可见
			var userIDs = make([]int32, 0)
			if err = json.Unmarshal([]byte(friend.PermissionUser), &userIDs); err != nil {
				fo.Log.Error("解析部分可见失败", zap.Error(err))
			}
			for _, userID := range userIDs {
				if userID == req.UserID {
					result = append(result, friend)
					req.Limit--
				}
			}

		case 6: //部分不可见
			var userIDs = make([]int32, 0)
			if err = json.Unmarshal([]byte(friend.PermissionUser), &userIDs); err != nil {
				fo.Log.Error("解析部分不可见失败", zap.Error(err))
			}
			for _, userID := range userIDs {
				if userID == req.UserID {
					continue
				}
			}
			result = append(result, friend)
			req.Limit--
		}

	}
	//

	//发帖人个人信息
	var beGetUser = []int32{}
	for i, forum := range result {
		beGetUser = append(beGetUser, result[i].CreateBy)
		//查点赞ThumbCheck
		var hasClickResp *moment.HasClickResp
		if hasClickResp, err = fo.ThumbDBServiceClient.HasClick(c.Request.Context(), &moment.Thumb{CreateBy: req.UserID, ForumID: forum.ForumID}); err != nil {
			fo.Response(c, "FORUM_FOLLOW", err, "检查是否点赞", nil)
			return
		}

		result[i].HasThumb = hasClickResp.State

		//查评论 先评论得在前面
		var commentPage *moment.CommentPageRep
		if commentPage, err = fo.CommentDBServiceClient.Page(c.Request.Context(), &moment.CommentPageReq{ForumID: forum.ForumID, Limit: req.CommentLimit}); err != nil {
			fo.Response(c, "FORUM_FOLLOW", err, "获取评论及回复", nil)
			return
		}

		//如果是推荐人  处理评论@
		for commentindex := range commentPage.Comments {
			if result[i].IsRecommend && commentPage.Comments[commentindex].AtEntity != "" {
				json.Unmarshal([]byte(commentPage.Comments[commentindex].AtEntity), &commentPage.Comments[commentindex].Entitys)
			}
		}

		result[i].Comments = commentPage.Comments
		//查点赞的人
		var thumbUserIDResp *moment.ThumbUserIDResp
		if thumbUserIDResp, err = fo.ThumbDBServiceClient.UserID(c.Request.Context(), &moment.ThumbPageReq{ForumID: forum.ForumID, Limit: req.ThumbLimit}); err != nil {
			fo.Response(c, "FORUM_FOLLOW", err, "获取点赞用户", nil)
			return
		}
		result[i].ThumbUsers = thumbUserIDResp.UserID

		//获取评论的用户信息
		for _, comment := range result[i].Comments {
			beGetUser = append(beGetUser, comment.CreateBy, comment.ReplayUID)
		}
		//查询点赞的用户信息
		beGetUser = append(beGetUser, result[i].ThumbUsers...)
		if result[i].ContentEntity != "" {
			json.Unmarshal([]byte(result[i].ContentEntity), &result[i].Entitys)
		}
	}

	var usersInfo *imapigateway.GetUserInfoByIdArrResult
	//查询点赞的用户信息
	if usersInfo, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, beGetUser, pkg.GetClaims(c).UserID); err != nil {
		fo.Response(c, "FORUM_FOLLOW", err, "获取IM用户信息", nil)
		return
	}
	//
	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i := range usersInfo.Uinfo {
		userInfoMapper[usersInfo.Uinfo[i].UserId] = usersInfo.Uinfo[i]
	}

	for i := range result {
		result[i].Creator = userInfoMapper[result[i].CreateBy]

		//获取评论的用户信息
		for j := range result[i].Comments {
			result[i].Comments[j].Creator = userInfoMapper[result[i].Comments[j].CreateBy]
			result[i].Comments[j].ReplayUser = userInfoMapper[result[i].Comments[j].ReplayUID]

			if result[i].Comments[j].AtEntity != "" {
				json.Unmarshal([]byte(result[i].Comments[j].AtEntity), &result[i].Comments[j].Entitys)
			}
		}
		for k := range result[i].ThumbUserInfos {
			result[i].ThumbUserInfos[k] = userInfoMapper[result[i].ThumbUsers[k]]
		}
	}
	fo.Response(c, "FORUM_FOLLOW", err, "获取关注帖子", result)
}

//ForumFriend 朋友帖子
// @Tags 帖子
// @Summary 朋友帖子
// @Produce  json
// @Accept  json
// @Param Limit  query int true "查询条数"
// @Param ForumID  query int true "上次浏览ID"
// @Param Authorization header string true "Token"
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/friend [get]
func (fo *ForumService) ForumFriend(c *gin.Context) {
	var req = &moment.ForumFriendPageReq{}
	var err error
	if err = c.BindQuery(req); err != nil {
		fo.Response(c, "FORUM_FRIEND", err, "解析朋友帖子参数", nil)
		return
	}
	req.UserID = pkg.GetClaims(c).UserID

	var fs = &moment.ForumFriendPageRep{}

	if fs, err = fo.ForumDBServiceClient.FriendPage(c.Request.Context(), req); err != nil {
		fo.Response(c, "FORUM_FRIEND", err, "获取朋友帖子", nil)
		return
	}

	for i, forum := range fs.Forums {

		//查询媒体
		if fs.Forums[i].ContentType > 1 {
			var mediaGetRep *moment.MediaGetRep
			if mediaGetRep, err = fo.MediaDBServiceClient.Get(c.Request.Context(), &moment.Media{MainID: forum.ForumID}); err != nil {
				fo.Response(c, "FORUM_FRIEND", err, "获取媒体", nil)
				return
			}
			fs.Forums[i].Medias = mediaGetRep.Medias
		}
		if req.UserID != forum.CreateBy {
			//是否关注
			var follow *moment.Follow
			if follow, err = fo.FollowDBServiceClient.Get(c.Request.Context(), &moment.Follow{CreateBy: req.UserID, FollowUID: forum.CreateBy}); err != nil {
				fo.Response(c, "FORUM_FRIEND", err, "检查是否关注", nil)
				return
			}
			fs.Forums[i].HasFollow = follow.FollowID > 0
		}
		//查点赞ThumbCheck
		var hasClickResp *moment.HasClickResp
		if hasClickResp, err = fo.ThumbDBServiceClient.HasClick(c.Request.Context(), &moment.Thumb{CreateBy: req.UserID, ForumID: forum.ForumID}); err != nil {
			fo.Response(c, "FORUM_FRIEND", err, "检查是否点赞", nil)
			return
		}

		fs.Forums[i].HasThumb = hasClickResp.State

		//查评论 先评论得在前面
		var commentPage *moment.CommentPageRep
		if commentPage, err = fo.CommentDBServiceClient.AllPage(c.Request.Context(), &moment.CommentPageReq{ForumID: forum.ForumID, Limit: 0}); err != nil {
			fo.Response(c, "FORUM_FRIEND", err, "获取评论及回复", nil)
			return
		}
		fs.Forums[i].Comments = commentPage.Comments
		//处理@
		json.Unmarshal([]byte(fs.Forums[i].ContentEntity), &fs.Forums[i].Entitys)
	}

	//发帖人个人信息
	var beGetUser = []int32{}
	for i := range fs.Forums {
		beGetUser = append(beGetUser, fs.Forums[i].CreateBy)
		//获取评论的用户信息
		for _, comment := range fs.Forums[i].Comments {
			beGetUser = append(beGetUser, comment.CreateBy, comment.ReplayUID)
		}
		//查询点赞的用户信息
	}

	var result *imapigateway.GetUserInfoByIdArrResult
	//查询点赞的用户信息
	if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, beGetUser, pkg.GetClaims(c).UserID); err != nil {
		fo.Response(c, "FORUM_RECOMMEND", err, "获取IM用户信息", nil)
		return
	}
	//
	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i := range result.Uinfo {
		userInfoMapper[result.Uinfo[i].UserId] = result.Uinfo[i]
	}

	for i := range fs.Forums {
		fs.Forums[i].Creator = userInfoMapper[fs.Forums[i].CreateBy]
		//获取评论的用户信息
		for j := range fs.Forums[i].Comments {
			fs.Forums[i].Comments[j].Creator = userInfoMapper[fs.Forums[i].Comments[j].CreateBy]
			fs.Forums[i].Comments[j].ReplayUser = userInfoMapper[fs.Forums[i].Comments[j].ReplayUID]
		}
		for k := range fs.Forums[i].ThumbUserInfos {
			fs.Forums[i].ThumbUserInfos[k] = userInfoMapper[fs.Forums[i].ThumbUsers[k]]
		}
	}

	fo.Response(c, "FORUM_TOPIC", err, "获取朋友帖子", fs.Forums)
}

type ForumOtherMainResp struct {
	Forums      []*moment.ForumFriend
	LastForumID int64
	Finish      bool
}

//ForumOtherMain 他人主页帖子
// @Tags 帖子
// @Summary 他人主页帖子
// @Produce  json
// @Accept  json
// @Param Limit  query int true "查询条数"
// @Param ForumID  query int true "上次浏览ID 取返回LastForumID"
// @Param Authorization header string true "Token"
// @Param FriendID  query int false "朋友ID"
// @Param RoundNum  query int false "查询轮次"
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/othermain [get]
func (fo *ForumService) ForumOtherMain(c *gin.Context) {
	var req = &moment.OtherMainPageReq{}
	var err error
	if err = c.BindQuery(req); err != nil {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "解析他人帖子参数", nil)
		return
	}
	if req.FriendID == 0 || req.RoundNum == 0 {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "FriendID、RoundNum不能为0", nil)
		return
	}
	req.UserID = pkg.GetClaims(c).UserID
	//检查是否设置了忽略
	var rep *moment.UserIgnoreCheckRep
	if rep, err = fo.BaseDBServiceClient.UserIgnoreCheck(c.Request.Context(), &moment.UserIgnore{UserID: req.UserID, CreateBy: req.FriendID}); err != nil {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "检查忽略设置失败", nil)
		return
	}
	if rep.Status {
		fo.Response(c, "FORUM_OTHER_MAIN_IGNORE", err, "获取他人帖子", nil)
		return
	}
	// //检查我和主人是否是好友
	var checkIsMuteContactResult *imapigateway.CheckIsMuteContactResult
	if checkIsMuteContactResult, err = pkg.CheckIsMuteContact(c.Request.Context(), fo.APIGatewayServiceClient, req.UserID, req.FriendID); err != nil {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "获取双向好友失败", err)
		return
	}

	var follow = &moment.Follow{CreateBy: req.UserID, FollowUID: req.FriendID}
	if follow, err = fo.FollowDBServiceClient.Get(c.Request.Context(), follow); err != nil {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "查询粉丝", err)
		return
	}

	// fo.Log.Debug("我的好友", zap.Any("DATA", req.Friends))
	// fs, err := internal.ForumTopic(fo.Cfg, req)

	var forumsReturn []*moment.ForumFriend
	var pageCount = req.Limit
	// for i := int32(1); len(forumsReturn) < int(pageNo); i++ {
	var fs = &moment.OtherMainPageRep{}
	var resp = &ForumOtherMainResp{}
	if req.Limit = req.RoundNum * req.Limit; req.Limit > 100 {
		req.Limit = 100
	}
	if fs, err = fo.ForumDBServiceClient.OtherMainPage(c.Request.Context(), req); err != nil {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "获取他人帖子", fs.Forums)
		return
	}

	if len(fs.Forums) < int(req.Limit) {
		resp.Finish = true
	}

	for i := range fs.Forums {
		if len(forumsReturn) == int(pageCount) {
			break
		}
		switch fs.Forums[i].Permission {
		case 1:
			forumsReturn = append(forumsReturn, fs.Forums[i])
			continue
		case 3: //粉丝
			if follow.ID > 0 {
				forumsReturn = append(forumsReturn, fs.Forums[i])
				continue
			}
		case 4: //朋友
			if checkIsMuteContactResult.Result {
				forumsReturn = append(forumsReturn, fs.Forums[i])
				continue
			}
		case 5: //给部分人
			if checkIsMuteContactResult.Result {
				var permissionUserIDs []int32
				if err = json.Unmarshal([]byte(fs.Forums[i].PermissionUser), &permissionUserIDs); err != nil {
					fo.Response(c, "FORUM_OTHER_MAIN", err, "解析部分可见人", fs.Forums)
					return
				}
				for _, permissionUserID := range permissionUserIDs {
					if permissionUserID == req.UserID {
						forumsReturn = append(forumsReturn, fs.Forums[i])
						continue
					}
				}
			}
		case 6: //不给
			if checkIsMuteContactResult.Result {
				var permissionUserIDs []int32
				if err = json.Unmarshal([]byte(fs.Forums[i].PermissionUser), &permissionUserIDs); err != nil {
					fo.Response(c, "FORUM_OTHER_MAIN", err, "解析部分不可见人", fs.Forums)
					return
				}
				for _, permissionUserID := range permissionUserIDs {
					if permissionUserID == req.UserID {
						continue
					}
				}
				forumsReturn = append(forumsReturn, fs.Forums[i])
				continue
			}
		}
		// }
	}

	if len(forumsReturn) > 0 {
		resp.LastForumID = forumsReturn[len(forumsReturn)-1].ForumID
	} else if len(fs.Forums) > 0 {
		resp.LastForumID = fs.Forums[len(fs.Forums)-1].ForumID
	}

	//发帖人个人信息
	var beGetUser = []int32{}
	for i := range forumsReturn {
		beGetUser = append(beGetUser, forumsReturn[i].CreateBy)
		//获取评论的用户信息
		for _, comment := range forumsReturn[i].Comments {
			beGetUser = append(beGetUser, comment.CreateBy, comment.ReplayUID)
		}
		//查询点赞的用户信息
		beGetUser = append(beGetUser, forumsReturn[i].ThumbUsers...)
		//处理高亮
		if forumsReturn[i].ContentEntity != "" {
			json.Unmarshal([]byte(forumsReturn[i].ContentEntity), &forumsReturn[i].Entitys)
		}
		//获取图片信息
		if forumsReturn[i].ContentType > 1 {
			var resp *moment.MediaGetRep
			if resp, err = fo.MediaDBServiceClient.Get(c.Request.Context(), &moment.Media{MainID: forumsReturn[i].ForumID}); err != nil {
				fo.Response(c, "FORUM_OTHER_MAIN", err, "获取媒体出错", nil)
				return
			}
			forumsReturn[i].Medias = resp.Medias
		}
	}

	var result *imapigateway.GetUserInfoByIdArrResult
	//查询点赞的用户信息
	if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, beGetUser, pkg.GetClaims(c).UserID); err != nil {
		fo.Response(c, "FORUM_OTHER_MAIN", err, "获取IM用户信息", nil)
		return
	}
	//
	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i := range result.Uinfo {
		userInfoMapper[result.Uinfo[i].UserId] = result.Uinfo[i]
	}

	for i := range forumsReturn {
		forumsReturn[i].Creator = userInfoMapper[forumsReturn[i].CreateBy]
		//获取评论的用户信息
		for j := range forumsReturn[i].Comments {
			forumsReturn[i].Comments[j].Creator = userInfoMapper[forumsReturn[i].Comments[j].CreateBy]
			forumsReturn[i].Comments[j].ReplayUser = userInfoMapper[forumsReturn[i].Comments[j].ReplayUID]
		}
		for k := range forumsReturn[i].ThumbUserInfos {
			forumsReturn[i].ThumbUserInfos[k] = userInfoMapper[forumsReturn[i].ThumbUsers[k]]
		}
	}
	resp.Forums = forumsReturn
	fo.Response(c, "FORUM_OTHER_MAIN", err, "获取他人帖子", resp)
}

//ForumSelfMain 话题帖子
// @Tags 帖子
// @Summary自己主页帖子
// @Produce  json
// @Accept  json
// @Param Limit  query int true "查询条数"
// @Param ForumID  query int true "上次浏览ID"
// @Param Authorization header string true "Token"
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /forumsvc/selfmain [get]
func (fo *ForumService) ForumSelfMain(c *gin.Context) {
	var req = &moment.SelfMainPageReq{}

	var err error
	if err = c.BindQuery(req); err != nil {
		fo.Response(c, "FORUM_SELF_MAIN", err, "解析我的帖子参数", nil)
		return
	}
	req.UserID = pkg.GetClaims(c).UserID
	// if req.Friends, err = pkg.GetDoubleDiffusionFriend(c.Request.Context(), fo.APIGatewayServiceClient, req.CreateBy); err != nil {
	// 	fo.Response(c, "FORUM_SELF_MAIN", err, "获取我的好友失败", err)
	// 	return
	// }
	// fo.Log.Debug("我的好友", zap.Any("DATA", req.Friends))
	var fs = &moment.SelfMainPageRep{}

	if fs, err = fo.ForumDBServiceClient.SelfMainPage(c.Request.Context(), req); err != nil {
		fo.Response(c, "FORUM_SELF_MAIN", err, "获取我的帖子", nil)
		return
	}
	//发帖人个人信息
	var beGetUser = []int32{}
	for i := range fs.Forums {
		beGetUser = append(beGetUser, fs.Forums[i].CreateBy)
		//获取评论的用户信息
		for _, comment := range fs.Forums[i].Comments {
			beGetUser = append(beGetUser, comment.CreateBy, comment.ReplayUID)
		}
		//查询点赞的用户信息
		beGetUser = append(beGetUser, fs.Forums[i].ThumbUsers...)
	}

	var result *imapigateway.GetUserInfoByIdArrResult
	//查询点赞的用户信息
	if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, beGetUser, pkg.GetClaims(c).UserID); err != nil {
		fo.Response(c, "FORUM_RECOMMEND", err, "获取IM用户信息", nil)
		return
	}
	//
	var userInfoMapper = make(map[int32]*imapigateway.UserInfo)
	for i := range result.Uinfo {
		userInfoMapper[result.Uinfo[i].UserId] = result.Uinfo[i]
	}

	for i := range fs.Forums {
		fs.Forums[i].Creator = userInfoMapper[fs.Forums[i].CreateBy]
		//获取评论的用户信息
		for j := range fs.Forums[i].Comments {
			fs.Forums[i].Comments[j].Creator = userInfoMapper[fs.Forums[i].Comments[j].CreateBy]
			fs.Forums[i].Comments[j].ReplayUser = userInfoMapper[fs.Forums[i].Comments[j].ReplayUID]
		}
		for k := range fs.Forums[i].ThumbUserInfos {
			fs.Forums[i].ThumbUserInfos[k] = userInfoMapper[fs.Forums[i].ThumbUsers[k]]
		}
	}

	fo.Response(c, "FORUM_SELF_MAIN", err, "获取我的帖子", fs.Forums)
}
