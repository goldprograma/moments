package svc

import (
	"context"
	"errors"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
	"strings"

	"go.uber.org/zap"
)

//Add 新增关注
func (fo *FollowService) Add(c context.Context, params *moment.Follow) (*moment.Follow, error) {
	var err error
	if params.FollowUID == 0 || params.CreateBy == 0 {
		return params, errors.New("参数检测出错FollowUID、CreateBy")
	}
	if params.CreateBy == params.FollowUID {
		return params, errors.New("不能自己关注自己")
	}

	params, err = fo.FollowDBServiceClient.Add(c, params)
	if err != nil && strings.Contains(err.Error(), "Error 1062") {
		err = errors.New("重复关注")
	}
	if err == nil {
		ctx := context.TODO()
		notic := &moment.Notice{Type: moment.NoticeType_Follow_Type, CreateBy: params.CreateBy, Notifier: params.FollowUID, Status: 1}
		if notic, err = fo.NoticeDBServiceClient.Get(ctx, notic); err != nil {
			fo.Log.Error("查询是否已关注过", zap.Error(err))
			return params, err
		}
		//已关注过 忽略二次关注
		if notic.ID > 0 {
			return params, err
		}
		if _, err = fo.NoticeDBServiceClient.Add(ctx, &moment.Notice{Type: moment.NoticeType_Follow_Type, CreateBy: params.CreateBy, Notifier: params.FollowUID, Status: 1}); err != nil {
			fo.Log.Error("新增关注消息出错", zap.Error(err))
			return params, err
		}
		// 发送互动消息   //TODO 个人版不需要发送互动消息
		// go func(params *moment.Follow) {
		// 	if err = pkg.ImMomentNotify(context.TODO(),
		// 		fo.APIGatewayServiceClient,
		// 		&imapigateway.MomentNotifyReq{
		// 			MomentNotifyData: []*imapigateway.MomentNotifyData{{
		// 				SourceCode:   4,
		// 				BusinessCode: 8,
		// 				ToId:         params.FollowUID,
		// 				WithId:       params.CreateBy,
		// 				IsFollowed:   int64(params.EachOther),
		// 				MsgTime:      params.CreateAt,
		// 			}}}); err != nil {
		// 		fo.Log.Error("发送IM消息出错", zap.Error(err))
		// 	}
		// }(params)
	}
	return params, err
}

//Delete 删除关注
func (fo *FollowService) Delete(c context.Context, params *moment.Follow) (*moment.Follow, error) {
	var err error
	if params.FollowUID == 0 || params.CreateBy == 0 {
		return params, errors.New("参数检测出错FollowUID、CreateBy")
	}
	params, err = fo.FollowDBServiceClient.Delete(c, params)

	return params, err
}

//FansCountBySource 删除关注
func (fo *FollowService) FansCountBySource(c context.Context, params *moment.Follow) (resp *moment.FansCountBySourceResp, err error) {
	if params.CreateBy == 0 || params.FollowSource == 0 {
		return resp, errors.New("参数检测出错FollowSource、CreateBy")
	}
	resp, err = fo.FollowDBServiceClient.FansCountBySourceInternal(c, params)

	return resp, err
}

//Get 获取关注
func (fo *FollowService) Get(c context.Context, params *moment.Follow) (resp *moment.Follow, err error) {
	if params.CreateBy == 0 || params.FollowUID == 0 {
		return resp, errors.New("参数检测出错FollowSource、FollowUID")
	}
	resp, err = fo.FollowDBServiceClient.Get(c, params)

	return resp, err
}

//FansID 分页获取粉丝ID
func (fo *FollowService) FansID(c context.Context, params *moment.FansIDReq) (resp *moment.FansIDResp, err error) {
	if params.UserID == 0 {
		return resp, errors.New("参数检测出错UserID")
	}
	resp, err = fo.FollowDBServiceClient.FansIDInternal(c, params)

	return resp, err
}

//Fans 粉丝分页
func (fo *FollowService) Fans(c context.Context, params *moment.FollowPageReq) (resp *moment.FollowPageRep, err error) {
	if params.CreateBy == 0 {
		return resp, errors.New("参数检测出错FollowSource、CreateBy")
	}
	if params.UserName != "" {
		//查询所有的关注ID 发送给IM IM
		if resp, err = fo.FollowDBServiceClient.FansAll(c, params); err != nil {
			fo.Log.Error("查询我的粉丝出错", zap.Error(err))
			return
		}
		var userIDs []int32
		var followMapper map[int32]*moment.Follow

		for _, follow := range resp.Follows {
			userIDs = append(userIDs, follow.CreateBy)
			followMapper[follow.CreateBy] = follow
		}
		//发送给IM
		//拿到userinfo
		var rep *imapigateway.SearchUserByNicknameResult
		if rep, err = fo.APIGatewayServiceClient.SearchUserByNickname(c,
			&imapigateway.SearchUserByNicknameReq{UserIds: userIDs, Nickname: params.UserName, Limit: int32(params.Limit), Offset: params.Offset}); err != nil {
			fo.Log.Error("模糊查询用户出错", zap.Error(err))
			return
		}
		resp.Follows = func() (follows []*moment.Follow) {
			for _, user := range rep.UserInfos {
				follows = append(follows, followMapper[user.UserId])
			}
			return
		}()
		resp.Users = rep.UserInfos
	} else {
		resp, err = fo.FollowDBServiceClient.Fans(c, params)
		var userIDs = make([]int32, 0, params.Limit)

		for _, follow := range resp.Follows {
			userIDs = append(userIDs, follow.CreateBy)
		}
		var result *imapigateway.GetUserInfoByIdArrResult
		if result, err = pkg.GetUserInfoByIDArr(c, fo.APIGatewayServiceClient, userIDs, params.CreateBy); err != nil {
			fo.Log.Error("查询用户信息出错", zap.Error(err))
			return
		}
		resp.Users = result.Uinfo
	}

	return resp, err
}

//Follows 我关注的分页
func (fo *FollowService) Follows(c context.Context, params *moment.FollowPageReq) (resp *moment.FollowPageRep, err error) {
	if params.CreateBy == 0 {
		return resp, errors.New("参数检测出错CreateBy")
	}

	if params.UserName != "" {
		//查询所有的关注ID 发送给IM IM
		if resp, err = fo.FollowDBServiceClient.FollowAll(c, params); err != nil {
			return
		}
		var userIDs []int32
		var followMapper = make(map[int32]*moment.Follow, 0)

		for _, follow := range resp.Follows {
			userIDs = append(userIDs, follow.FollowUID)
			followMapper[follow.FollowUID] = follow
		}
		//发送给IM
		//拿到userinfo
		var rep *imapigateway.SearchUserByNicknameResult
		if rep, err = fo.APIGatewayServiceClient.SearchUserByNickname(c,
			&imapigateway.SearchUserByNicknameReq{UserIds: userIDs, Nickname: params.UserName, Limit: params.Limit, Offset: params.Offset, SelfId: params.CreateBy}); err != nil {
			return
		}
		resp.Follows = func() (follows []*moment.Follow) {
			for _, user := range rep.UserInfos {
				follows = append(follows, followMapper[user.UserId])
			}
			return
		}()
		resp.Users = rep.UserInfos
	} else {

		resp, err = fo.FollowDBServiceClient.Follows(c, params)
		var userIDs = make([]int32, 0, params.Limit)

		for _, follow := range resp.Follows {
			userIDs = append(userIDs, follow.FollowUID)
		}

		var result *imapigateway.GetUserInfoByIdArrResult
		if result, err = pkg.GetUserInfoByIDArr(c, fo.APIGatewayServiceClient, userIDs, params.CreateBy); err != nil {
			return
		}
		resp.Users = result.Uinfo
	}

	return resp, err
}

//FansCount 我得粉丝数
func (fo *FollowService) FansCount(c context.Context, params *moment.FollowCountReq) (resp *moment.FollowCountRep, err error) {
	if params.UserID == 0 {
		return resp, errors.New("参数检测出错UserID")
	}
	resp, err = fo.FollowDBServiceClient.FansCount(c, params)

	return resp, err
}
