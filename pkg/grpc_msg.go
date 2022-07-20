package pkg

import (
	"errors"
	"log"

	"gitlab.moments.im/pkg/protoc/imapigateway"

	"golang.org/x/net/context"
)

//ImMomentNotify 通知IM 互动消息
func ImMomentNotify(ctx context.Context, apiGatewayServiceClient imapigateway.ApiGatewayServiceClient, msg *imapigateway.MomentNotifyReq) error {
	result, err := apiGatewayServiceClient.MomentNotify(ctx, msg)
	if err != nil {
		return err
	}
	if result.ErrorCode != 0 {
		return errors.New(result.ErrorMsg)
	}
	return nil
}

func UpdateMoments(ctx context.Context, ContactPushServiceClient imapigateway.ApiGatewayServiceClient, momentStates []*imapigateway.MomentState, selfId int32) (err error) {
	_, err = ContactPushServiceClient.UpdateMoments(ctx, &imapigateway.ReqUpdateMoments{MomentStates: momentStates, SelfId: selfId})
	return
}

func GetUserInfoByIDArr(ctx context.Context, apiGatewayServiceClient imapigateway.ApiGatewayServiceClient, userIDs []int32, selfID int32) (rep *imapigateway.GetUserInfoByIdArrResult, err error) {
	if rep, err = apiGatewayServiceClient.GetUserInfoByIdArr(ctx, &imapigateway.GetUserInfoByIdArrReq{SelfId: selfID, UserIds: userIDs}); err != nil {
		return nil, err
	}
	return
}

func GetUserInfoByID(ctx context.Context, apiGatewayServiceClient imapigateway.ApiGatewayServiceClient, selfID, userID int32) (rep *imapigateway.GetUserInfoResult, err error) {
	//
	log.Printf("self_id:%v\tuser_id:%v\n", selfID, userID)
	//
	if rep, err = apiGatewayServiceClient.GetUserInfoById(ctx, &imapigateway.GetUserInfoByIdReq{SelfId: selfID, UserId: userID}); err != nil {
		return nil, err
	}
	return
}

//CheckIsMuteContact 检查是不是双向好友
func CheckIsMuteContact(ctx context.Context, apiGatewayServiceClient imapigateway.ApiGatewayServiceClient, selfID, userID int32) (rep *imapigateway.CheckIsMuteContactResult, err error) {
	if rep, err = apiGatewayServiceClient.CheckIsMuteContact(ctx, &imapigateway.CheckIsMuteContactReq{SelfId: userID, UserId: userID}); err != nil {
		return nil, err
	}
	return
}

//GetCommonFriends 获取共同好友
func GetCommonFriends(ctx context.Context, apiGatewayServiceClient imapigateway.ApiGatewayServiceClient, selfID int32, friendid int32) (ids []int32, err error) {

	//获取IM用户好友
	var imContactsReq *imapigateway.ContactsGetCommonContactsResult
	if imContactsReq, err = apiGatewayServiceClient.ContactsGetCommonContacts(ctx, &imapigateway.ContactsGetCommonContactsReq{UserIds: []int32{selfID, friendid}}); err != nil {

		return nil, err
	}
	for _, user := range imContactsReq.UserInfos {
		if user.MutalContact {
			ids = append(ids, user.UserId)
		}
	}
	return
}

//GetDoubleDiffusionFriend 获取用户可扩散双向的好友
func GetDoubleDiffusionFriend(ctx context.Context, apiGatewayServiceClient imapigateway.ApiGatewayServiceClient, selfID int32) (ids []int32, err error) {

	//获取IM用户好友
	var imContactsReq *imapigateway.ContatsGetContactsResult
	if imContactsReq, err = apiGatewayServiceClient.ContactsGetContacts(ctx, &imapigateway.ContatsGetContactsReq{UserId: selfID}); err != nil {

		return nil, err
	}
	for _, user := range imContactsReq.UserInfoArr {
		if user.MutalContact {
			ids = append(ids, user.UserId)
		}
	}

	return
}
