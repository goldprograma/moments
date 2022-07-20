package svc

import (
	"context"
	"database/sql"
	"time"

	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type FollowDBService struct {
	pkg.BaseComponent
}

// Add 新增follow
func (fs *FollowDBService) Add(c context.Context, follow *moment.Follow) (*moment.Follow, error) {
	follow.FollowID = fs.GetID()
	follow.CreateAt = time.Now().Unix()

	has, err := internal.FollowCheck(follow.FollowUID, follow.CreateBy, fs.DB)
	if err != nil {
		return nil, err
	}
	if has {
		follow.EachOther = 1
	}
	var tx *sqlx.Tx
	if tx, err = fs.DB.BeginTxx(c, &sql.TxOptions{ReadOnly: false}); err != nil {

	}
	err = internal.FollowInsert(follow, fs.DB, tx)
	if err != nil {
		return nil, err
	}
	if has {
		if err = internal.FollowUpdate(&moment.Follow{FollowUID: follow.CreateBy, CreateBy: follow.FollowUID, EachOther: 1}, fs.DB, tx, "follow_uid", "create_by"); err != nil {
			return nil, err
		}
	}
	if err == nil {
		tx.Commit()
	}

	if err = internal.UserStatusFansCountUpdateDB(&moment.UserStatus{UserID: follow.FollowUID, FansCount: 1}, fs.DB); err != nil {
		fs.Log.Error("更新用户DB粉丝数失败", zap.Int32("用户", follow.FollowUID), zap.Error(err))
	}
	if err = internal.UserStatusFollowCountUpdateDB(&moment.UserStatus{UserID: follow.CreateBy, FollowCount: 1}, fs.DB); err != nil {
		fs.Log.Error("更新用户DB关注数失败", zap.Int32("用户", follow.CreateBy), zap.Error(err))
	}

	//更新用户点赞数
	if err = cache.UserStatusUpdateCache(&moment.UserStatus{UserID: follow.FollowUID, FansCount: 1}, fs.Redis); err != nil {
		fs.Log.Error("更新用户粉丝数失败", zap.Int32("用户", follow.FollowUID), zap.Error(err))
	}
	if err = cache.UserStatusUpdateCache(&moment.UserStatus{UserID: follow.CreateBy, FollowCount: 1}, fs.Redis); err != nil {
		fs.Log.Error("更新用户关注数失败", zap.Int32("用户", follow.CreateBy), zap.Error(err))
	}

	return follow, err
}

// Delete 删除follow
func (fs *FollowDBService) Delete(c context.Context, follow *moment.Follow) (*moment.Follow, error) {
	var tx *sqlx.Tx
	var err error
	if tx, err = fs.DB.BeginTxx(c, &sql.TxOptions{ReadOnly: false}); err != nil {
		return nil, err
	}
	if err = internal.FollowDelete(follow, fs.DB, tx); err != nil {
		return nil, err
	}
	if err = internal.FollowUpdate(&moment.Follow{FollowUID: follow.CreateBy, CreateBy: follow.FollowUID, EachOther: 0}, nil, tx, "follow_uid", "create_by"); err == nil {
		tx.Commit()
	}

	//更新用户关注数
	if err = internal.UserStatusFansCountUpdateDB(&moment.UserStatus{UserID: follow.FollowUID, FansCount: -1}, fs.DB); err != nil {
		fs.Log.Error("更新用户DB粉丝数失败", zap.Int32("用户", follow.FollowUID), zap.Error(err))
	}
	if err = internal.UserStatusFollowCountUpdateDB(&moment.UserStatus{UserID: follow.CreateBy, FollowCount: -1}, fs.DB); err != nil {
		fs.Log.Error("更新用户DB关注数失败", zap.Int32("用户", follow.CreateBy), zap.Error(err))
	}

	if err = cache.UserStatusUpdateCache(&moment.UserStatus{UserID: follow.FollowUID, FansCount: -1}, fs.Redis); err != nil {
		fs.Log.Error("更新用户粉丝数失败", zap.Int32("用户", follow.FollowUID), zap.Error(err))
	}
	if err = cache.UserStatusUpdateCache(&moment.UserStatus{UserID: follow.CreateBy, FollowCount: -1}, fs.Redis); err != nil {
		fs.Log.Error("更新用户关注数失败", zap.Int32("用户", follow.CreateBy), zap.Error(err))
	}

	return follow, err
}

// Get 获取某个关注
func (fs *FollowDBService) Get(c context.Context, follow *moment.Follow) (*moment.Follow, error) {
	err := internal.FollowGet(follow, fs.DB)
	return follow, err
}

// FollowAllOrderByCreateAt 关注根据关注时间排序
func (fs *FollowDBService) FollowAllOrderByCreateAt(c context.Context, req *moment.FollowAllOrderByCreateAtReq) (*moment.FollowAllOrderByCreateAtResp, error) {
	forums, err := internal.FollowAllOrderByCreateAt(req.UserID, req.CreateAt, fs.DB)
	return &moment.FollowAllOrderByCreateAtResp{Follows: forums}, err
}

//FollowCount 关注数量
func (fs *FollowDBService) FollowCount(c context.Context, follow *moment.FollowCountReq) (*moment.FollowCountRep, error) {
	count, err := internal.FollowCount(follow, fs.DB)
	return &moment.FollowCountRep{Count: count}, err
}

//FansCount 关注数量
func (fs *FollowDBService) FansCount(c context.Context, follow *moment.FollowCountReq) (*moment.FollowCountRep, error) {
	count, err := internal.FollowFansCount(follow, fs.DB)
	return &moment.FollowCountRep{Count: count}, err
}

// Follows 获取分页关注
func (fs *FollowDBService) Follows(c context.Context, page *moment.FollowPageReq) (*moment.FollowPageRep, error) {
	data, err := internal.FollowByMe(page, fs.DB)
	return &moment.FollowPageRep{Follows: data}, err
}

//Fans 关注我的
func (fs *FollowDBService) Fans(c context.Context, page *moment.FollowPageReq) (*moment.FollowPageRep, error) {
	data, err := internal.FollowMe(page, fs.DB)
	return &moment.FollowPageRep{Follows: data}, err
}

//FansAll 我的所有粉丝
func (fs *FollowDBService) FansAll(c context.Context, req *moment.FollowPageReq) (*moment.FollowPageRep, error) {
	data, err := internal.FansAll(req.CreateBy, fs.DB)
	return &moment.FollowPageRep{Follows: data}, err
}

//FansAllID 我的所有粉丝
func (fs *FollowDBService) FansAllID(c context.Context, req *moment.Follow) (*moment.FansAllIDRep, error) {
	data, err := internal.FollowMeAllID(req.FollowUID, fs.DB)
	return &moment.FansAllIDRep{Fans: data}, err
}

//FollowAllID 我的所有粉丝
func (fs *FollowDBService) FollowAllID(c context.Context, req *moment.Follow) (*moment.FollowAllIDRep, error) {
	data, err := internal.FollowAllID(req.FollowUID, fs.DB)
	return &moment.FollowAllIDRep{Fans: data}, err
}

//FollowAll 我的所有关注
func (fs *FollowDBService) FollowAll(c context.Context, req *moment.FollowPageReq) (*moment.FollowPageRep, error) {
	data, err := internal.FollowAll(req.CreateBy, fs.DB)
	return &moment.FollowPageRep{Follows: data}, err
}

//FansCountBySourceInternal 我的所有关注
func (fs *FollowDBService) FansCountBySourceInternal(c context.Context, req *moment.Follow) (*moment.FansCountBySourceResp, error) {
	data, err := internal.FansCountBySource(req.CreateBy, req.FollowSource, fs.DB)
	return &moment.FansCountBySourceResp{Count: data}, err
}

//FansIDInternal 我的所有关注
func (fs *FollowDBService) FansIDInternal(c context.Context, req *moment.FansIDReq) (*moment.FansIDResp, error) {
	data, err := internal.FollowFansIDPage(req, fs.DB)
	return &moment.FansIDResp{UID: data}, err
}
