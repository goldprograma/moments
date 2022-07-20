package svc

import (
	"context"
	"errors"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type FollowService struct {
	pkg.BaseComponent
	FollowDBServiceClient   moment.FollowDBServiceClient
	NoticeDBServiceClient   moment.NoticeDBServiceClient
	APIGatewayServiceClient imapigateway.ApiGatewayServiceClient
}

//FollowAdd 新增关注
// @Tags 关注
// @Summary 新增关注
// @Produce  json
// @Description 只传 FollowUID
// @Param body body moment.Follow true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/add [post]
func (fo *FollowService) FollowAdd(c *gin.Context) {
	var params = &moment.Follow{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		fo.Response(c, "FOLLOW_ADD", err, "解析新增关注参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID

	if params.FollowUID == 0 {
		fo.Response(c, "FOLLOW_ADD", errors.New("参数检测出错FollowUID"), "解析新增关注参数", nil)
		return
	}
	if params.CreateBy == params.FollowUID {
		fo.Response(c, "FOLLOW_ADD", errors.New("不能自己关注自己"), "不能自己关注自己", nil)
		return
	}
	params, err = fo.FollowDBServiceClient.Add(c.Request.Context(), params)
	if err != nil && strings.Contains(err.Error(), "Error 1062") {
		err = errors.New("重复关注")
	}
	if err == nil {

		//发送互动消息
		go func(params *moment.Follow) {
			ctx := context.TODO()
			notic := &moment.Notice{Type: moment.NoticeType_Follow_Type, CreateBy: params.CreateBy, Notifier: params.FollowUID, Status: 1}
			if notic, err = fo.NoticeDBServiceClient.Get(ctx, notic); err != nil {
				fo.Log.Error("查询是否已关注过", zap.Error(err))
				return
			}
			//已关注过 忽略二次关注
			if notic.ID > 0 {
				return
			}
			if _, err = fo.NoticeDBServiceClient.Add(ctx, &moment.Notice{Type: moment.NoticeType_Follow_Type, CreateBy: params.CreateBy, Notifier: params.FollowUID, Status: 1}); err != nil {
				fo.Log.Error("新增关注消息出错", zap.Error(err))
				return
			}

			if err = pkg.ImMomentNotify(context.TODO(),
				fo.APIGatewayServiceClient,
				&imapigateway.MomentNotifyReq{
					MomentNotifyData: []*imapigateway.MomentNotifyData{{
						SourceCode:   4,
						BusinessCode: 8,
						ToId:         params.FollowUID,
						WithId:       params.CreateBy,
						IsFollowed: func() int64 {
							if int64(params.EachOther) == 1 {
								return 1
							}
							return 2
						}(),
						MsgTime: params.CreateAt,
					}}}); err != nil {
				fo.Log.Error("发送IM消息出错", zap.Error(err))
			}
		}(params)
	}
	fo.Response(c, "FOLLOW_ADD", err, "新增关注", params)
}

//FollowDelete 取消关注
// @Tags 关注
// @Summary 取消关注
// @Description 只传 FollowUID
// @Produce  json
// @Description 只传 FollowUID
// @Param body body moment.Follow true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/delete [post]
func (fo *FollowService) FollowDelete(c *gin.Context) {
	var params = &moment.Follow{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		fo.Response(c, "FOLLOW_DELETE", err, "解析取消关注参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID

	if params.FollowUID == 0 {
		fo.Response(c, "FOLLOW_DELETE", errors.New("参数检测出错FollowUID"), "解析取消关注", nil)
		return
	}

	params, err = fo.FollowDBServiceClient.Delete(c.Request.Context(), params)

	fo.Response(c, "FOLLOW_DELETE", err, "取消关注", params)
}

//FollowCheck 检查是否关注
// @Tags 关注
// @Summary 检查是否关注
// @Produce  json
// @Param FollowUID  query int true "被关注人ID"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/check [get]
func (fo *FollowService) FollowCheck(c *gin.Context) {
	var params = &moment.Follow{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		fo.Response(c, "FOLLOW_CHECK", err, "解析获取关注参数", nil)
		return
	}

	params.CreateBy = pkg.GetClaims(c).UserID

	if params.FollowUID == 0 {
		fo.Response(c, "FOLLOW_CHECK", errors.New("参数检测出错FollowUID"), "解析获取关注", nil)
		return
	}
	var result bool
	params, err = fo.FollowDBServiceClient.Get(c.Request.Context(), params)
	if params.ID > 0 {
		result = true
	}
	fo.Response(c, "FOLLOW_CHECK", err, "获取关注", result)
}

//FollowGet 获取关注
// @Tags 关注
// @Summary 获取关注
// @Produce  json
// @Param FollowUID  query int true "被关注人ID"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/get [get]
func (fo *FollowService) FollowGet(c *gin.Context) {
	var params = &moment.Follow{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		fo.Response(c, "FOLLOW_GET", err, "解析获取关注参数", nil)
		return
	}

	params.CreateBy = pkg.GetClaims(c).UserID

	if params.FollowUID == 0 {
		fo.Response(c, "FOLLOW_GET", errors.New("参数检测出错FollowUID"), "解析获取关注", nil)
		return
	}

	params, err = fo.FollowDBServiceClient.Get(c.Request.Context(), params)
	fo.Response(c, "FOLLOW_GET", err, "获取关注", params)
}

//HTTPFans 所有关注
// @Tags 关注
// @Summary 我的粉丝
// @Param Offset  query int false "偏移量"
// @Param Limit  query int false "分页大小"
// @Param UserName  query string false "用户昵称"
// @Param Authorization header string true "Token"
// @Produce  json
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/fans [get]
func (fo *FollowService) HTTPFans(c *gin.Context) {
	var params = &moment.FollowPageReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		fo.Response(c, "FOLLOW_FANS", err, "解析我的粉丝参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID
	var data *moment.FollowPageRep
	if params.UserName != "" {
		//查询所有的关注ID 发送给IM IM
		if data, err = fo.FollowDBServiceClient.FansAll(c.Request.Context(), params); err != nil {
			fo.Response(c, "FOLLOW_FANS", err, "查询我的粉丝出错", nil)
			return
		}
		var userIDs []int32
		var followMapper map[int32]*moment.Follow

		for _, follow := range data.Follows {
			userIDs = append(userIDs, follow.CreateBy)
			followMapper[follow.CreateBy] = follow
		}
		//发送给IM
		//拿到userinfo
		var rep *imapigateway.SearchUserByNicknameResult
		if rep, err = fo.APIGatewayServiceClient.SearchUserByNickname(c.Request.Context(),
			&imapigateway.SearchUserByNicknameReq{UserIds: userIDs, Nickname: params.UserName, Limit: params.Limit, Offset: params.Offset}); err != nil {
			fo.Response(c, "FOLLOW_FANS", err, "模糊查询用户", nil)
			return
		}
		data.Follows = func() (follows []*moment.Follow) {
			for _, user := range rep.UserInfos {
				follows = append(follows, followMapper[user.UserId])
			}
			return
		}()
		data.Users = rep.UserInfos
	} else {
		data, err = fo.FollowDBServiceClient.Fans(c.Request.Context(), params)
		var userIDs = make([]int32, 0, params.Limit)

		for _, follow := range data.Follows {
			userIDs = append(userIDs, follow.CreateBy)
		}
		var result *imapigateway.GetUserInfoByIdArrResult
		if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, userIDs, params.CreateBy); err != nil {
			fo.Response(c, "FOLLOW_FANS", err, "查询用户信息", nil)
			return
		}
		data.Users = result.Uinfo
	}
	fo.Response(c, "FOLLOW_FANS", err, "获取我的粉丝", data)
}

//HTTPFansCount 我的粉丝数量
// @Tags 关注
// @Summary 我的粉丝数量
// @Produce  json
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/fanscount [get]
func (fo *FollowService) HTTPFansCount(c *gin.Context) {
	var params = &moment.FollowCountReq{UserID: pkg.GetClaims(c).UserID}
	var err error

	var data *moment.FollowCountRep
	data, err = fo.FollowDBServiceClient.FansCount(c.Request.Context(), params)
	fo.Response(c, "FOLLOW_FANS_COUNT", err, "解析我的关注参数", data)

}

//Me 我关注的
// @Tags 关注
// @Summary 我关注的
// @Param Offset  query int false "偏移量"
// @Param Limit  query int false "分页大小"
// @Param UserName  query string false "用户昵称"
// @Param Authorization header string true "Token"
// @Produce  json
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/me [get]
func (fo *FollowService) Me(c *gin.Context) {
	var params = &moment.FollowPageReq{}
	var err error
	if err = c.BindQuery(params); err != nil {
		fo.Response(c, "FOLLOW_BY_ME", err, "解析查询用户信息", nil)
		return
	}
	var data *moment.FollowPageRep
	params.CreateBy = pkg.GetClaims(c).UserID

	if params.UserName != "" {
		//查询所有的关注ID 发送给IM IM
		if data, err = fo.FollowDBServiceClient.FollowAll(c.Request.Context(), params); err != nil {
			fo.Response(c, "FOLLOW_FANS", err, "查询我的粉丝出错", nil)
			return
		}
		var userIDs []int32
		var followMapper = make(map[int32]*moment.Follow, 0)

		for _, follow := range data.Follows {
			userIDs = append(userIDs, follow.FollowUID)
			followMapper[follow.FollowUID] = follow
		}
		//发送给IM
		//拿到userinfo
		var rep *imapigateway.SearchUserByNicknameResult
		if rep, err = fo.APIGatewayServiceClient.SearchUserByNickname(c.Request.Context(),
			&imapigateway.SearchUserByNicknameReq{UserIds: userIDs, Nickname: params.UserName, Limit: params.Limit, Offset: params.Offset, SelfId: params.CreateBy}); err != nil {
			fo.Response(c, "FOLLOW_FANS", err, "模糊查询用户", nil)
			return
		}
		data.Follows = func() (follows []*moment.Follow) {
			for _, user := range rep.UserInfos {
				follows = append(follows, followMapper[user.UserId])
			}
			return
		}()
		data.Users = rep.UserInfos
	} else {

		data, err = fo.FollowDBServiceClient.Follows(c.Request.Context(), params)
		var userIDs = make([]int32, 0, params.Limit)

		for _, follow := range data.Follows {
			userIDs = append(userIDs, follow.FollowUID)
		}

		var result *imapigateway.GetUserInfoByIdArrResult
		if result, err = pkg.GetUserInfoByIDArr(c.Request.Context(), fo.APIGatewayServiceClient, userIDs, params.CreateBy); err != nil {
			fo.Response(c, "FOLLOW_BY_ME", err, "查询用户信息", nil)
			return
		}
		data.Users = result.Uinfo
	}
	fo.Response(c, "FOLLOW_BY_ME", err, "获取我的关注", data)
}

//MeCount 我关注人的数量
// @Tags 关注
// @Summary 我关注人的数量
// @Param Authorization header string true "Token"
// @Produce  json
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /followsvc/mecount [get]
func (fo *FollowService) MeCount(c *gin.Context) {
	var params = &moment.FollowCountReq{UserID: pkg.GetClaims(c).UserID}
	var err error
	var rep *moment.FollowCountRep
	rep, err = fo.FollowDBServiceClient.FollowCount(c.Request.Context(), params)
	fo.Response(c, "FOLLOW_BY_ME_COUNT", err, "获取我关注的数量", rep)

}
