package svc

import (
	"context"
	"moments/pkg/protoc/moment"

	"go.uber.org/zap"
)

type RetryStrc struct {
	BusCode    string
	Data       interface{}
	RetryCount int32
}

//RetryChan 重试Channel
var RetryChan = make(chan RetryStrc, 1024)

func (bs *BaseService) UserAlbums(c context.Context, req *moment.UserAlbumsReq) (resp *moment.UserAlbumsRep, err error) {
	bs.Log.Debug("UserAlbums GRPC Req ", zap.Any("req", req))
	resp = &moment.UserAlbumsRep{}
	//直接查询公开的最新
	var userAlbumRep *moment.UserAlbumRep
	if userAlbumRep, err = bs.BaseDBServiceClient.UserAlbumFromCache(c, &moment.UserAlbumReq{UserID: req.UserID, FriendID: req.FriendID}); err != nil {
		return
	}
	bs.Log.Debug("UserAlbums GRPC resp ", zap.Any("resp", userAlbumRep.Medias))
	resp.Medias = userAlbumRep.Medias
	resp.Code = 1
	return
}

//UserContactsSync 联系人变动通知
//1.删除帖子关系
//2.删除关注
//3.减少点赞量(userstate or select count(*))
//ContactID 1,2,3,4
func (bs *BaseService) UserContactsSync(ctx context.Context, req *moment.UserContactsSyncReq) (*moment.UserContactsSyncRep, error) {
	var err error
	if req.Action < 0 || req.UserID < 1 || len(req.ContactID) == 0 {
		return &moment.UserContactsSyncRep{Code: 0, Msg: "参数缺失"}, err
	}
	bs.Log.Debug("收到IM通知", zap.Int32("用户", req.GetUserID()), zap.Int32("Action", req.GetAction()), zap.Any("ContactID", req.GetContactID()), zap.Error(err))

	go func(ctx context.Context, req *moment.UserContactsSyncReq) {
		if req.Action > 1 { //只要拉黑和删除
			for _, id := range req.ContactID {
				if _, err = bs.ForumDBServiceClient.UserDelete(ctx, &moment.ForumFriend{UserID: id, CreateBy: req.UserID}); err != nil {
					bs.Log.Debug("删除我的信箱失败,重新尝试", zap.Int32("用户", req.GetUserID()), zap.Int32("Action", req.GetAction()), zap.Int32("CreateBy", id), zap.Error(err))
					req.ContactID = append(req.ContactID, id)
					RetryChan <- RetryStrc{BusCode: "UserContactsSync", Data: &moment.ForumFriend{UserID: id, CreateBy: req.UserID}}
				}
			}
		}
	}(ctx, req)

	return &moment.UserContactsSyncRep{Code: 1, Msg: "联系人变动通知处理成功"}, err
}

func (bs *BaseService) Retry() {
	var err error
	bs.Log.Debug("开启重试协程")
	for {
		select {
		case s := <-RetryChan:
			switch s.BusCode {
			case "UserContactsSync":
				if s.RetryCount > 0 {
					s.RetryCount--
					forum := s.Data.(*moment.ForumFriend)
					if _, err = bs.ForumDBServiceClient.UserDelete(context.TODO(), &moment.ForumFriend{UserID: forum.CreateBy, CreateBy: forum.UserID}); err != nil {
						bs.Log.Debug("UserContactsSync删除我的信箱失败,重新尝试", zap.Int32("用户", forum.GetUserID()), zap.Int32("CreateBy", forum.GetCreateBy()), zap.Int32("剩余重试次数", s.RetryCount), zap.Error(err))
						RetryChan <- RetryStrc{BusCode: "UserContactsSync", Data: forum, RetryCount: s.RetryCount}
					}
				}
			}
		}
	}
}
