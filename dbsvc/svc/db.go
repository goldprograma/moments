package svc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

type BaseDBService struct {
	pkg.BaseComponent
}

//TagAdd 新增Tag
func (db *BaseDBService) TagAdd(ctx context.Context, req *moment.TagAddReq) (rep *moment.TagAddRep, err error) {
	req.Tag.TagID = db.GetID()
	req.Tag.CreateAt = time.Now().Unix()
	var tx *sqlx.Tx
	if tx, err = db.DB.BeginTxx(context.TODO(), &sql.TxOptions{ReadOnly: false}); err != nil {
		return nil, err
	}
	if err = internal.TagInsert(req.Tag, db.DB, tx); err != nil {
		tx.Rollback()
		return nil, err
	}
	for i := range req.UserTags {
		req.UserTags[i].TagID = req.Tag.TagID
		req.UserTags[i].UserTagID = db.GetID()
	}
	if err = internal.UserTagInsert(req.UserTags, db.DB, tx); err != nil {
		tx.Rollback()
	}
	tx.Commit()

	return &moment.TagAddRep{Tag: req.Tag}, err
}

//TagGet 新增Tag
func (db *BaseDBService) TagGet(ctx context.Context, req *moment.TagGetReq) (rep *moment.TagGetRep, err error) {
	tags, err := internal.TagGet(req.UserID, db.DB)
	if err != nil {
		return nil, err
	}
	var userTags = make([]*moment.UserTags, 0)
	for _, tag := range tags {
		var userTag []*moment.UserTag
		userTag, err = internal.UserTagGet(req.UserID, tag.TagID, db.DB)
		userTags = append(userTags, &moment.UserTags{Tag: tag, TagUsers: userTag})
	}

	return &moment.TagGetRep{UserTags: userTags}, err
}

//UserIgnoreGet 获取忽略配置
func (db *BaseDBService) UserIgnoreGet(ctx context.Context, req *moment.UserIgnore) (result *moment.IgnoreSlice, err error) {
	result = &moment.IgnoreSlice{}
	if req.CreateBy == 0 {
		return nil, errors.New("CreateBy")
	}
	result.Ignores, err = internal.UserIgnoreGet(req, db.DB)
	return result, err
}

//UserIgnoreCheck 检查忽略设置
func (db *BaseDBService) UserIgnoreCheck(ctx context.Context, req *moment.UserIgnore) (*moment.UserIgnoreCheckRep, error) {
	var err error
	if req.CreateBy == 0 || req.UserID == 0 {
		return nil, errors.New("CreateBy、UserID不能为空")
	}
	var status bool
	status, err = internal.UserIgnoreCheck(req, db.DB)
	return &moment.UserIgnoreCheckRep{Status: status}, err
}

//UserIgnoreAdd 获取忽略配置
func (db *BaseDBService) UserIgnoreAdd(ctx context.Context, req *moment.IgnoreSlice) (rep *moment.IgnoreNone, err error) {
	if len(req.Ignores) == 0 {
		return nil, errors.New("无新增忽略")
	}
	for i := range req.Ignores {
		if req.Ignores[i].UserID == 0 || req.Ignores[i].CreateBy == 0 {
			return nil, fmt.Errorf("第%d个元素UserID、CreateBy缺失", i)
		}
		req.Ignores[i].IgnoreID = db.GetID()
		req.Ignores[i].CreateAt = time.Now().Unix()
	}
	err = internal.UserIgnoreAdd(req.Ignores, db.DB)
	return &moment.IgnoreNone{}, err
}

//UserIgnoreDelete 获取忽略配置
func (db *BaseDBService) UserIgnoreDelete(ctx context.Context, req *moment.IgnoreSlice) (*moment.IgnoreNone, error) {
	if len(req.Ignores) == 0 {
		return nil, errors.New("CreateBy、UserID、UserIgnoreID不能为空")
	}
	var err error
	for _, ignore := range req.Ignores {
		if err = internal.UserIgnoreDelete(ignore, db.DB); err != nil {
			return nil, err
		}
	}
	return &moment.IgnoreNone{}, err
}

//UserStatisticsGet 获取用户朋友圈统计数据
func (db *BaseDBService) UserStatisticsGet(ctx context.Context, req *moment.UserStatus) (*moment.UserStatus, error) {
	if req.UserID == 0 {
		return nil, errors.New("UserID不能为空")
	}

	// err := cache.UserStatisticsCacheGet(req, db.Log, db.Redis, db.DB)
	userStatus, err := cache.GetUserStatusByUserID(db.DB, db.Redis, db.Log, int(req.UserID))
	return &userStatus, err
}

//UserHomeBackgroudGet 获取用户统计忽略配置
func (db *BaseDBService) UserHomeBackgroudGet(ctx context.Context, req *moment.UserStatus) (*moment.UserStatus, error) {
	if req.UserID == 0 {
		return nil, errors.New("UserID不能为空")
	}

	err := internal.UserHomeBackgroudGet(req, db.DB)
	return req, err
}

//UserVersionGet 获取用户版本(实际上起到初始化user_status表的作用)
func (db *BaseDBService) UserVersionGet(ctx context.Context, req *moment.UserStatus) (resp *moment.UserVersionGetRep, err error) {
	if req.UserID == 0 {
		return nil, errors.New("UserID不能为空")
	}
	err = cache.InitUserStatus(db.DB, int(req.UserID), db.Log)
	if err != nil {
		db.Log.Error("BaseDBService UserVersionGet cache.InitUserStatus => " + err.Error())
	}
	resp = &moment.UserVersionGetRep{}

	// err := cache.UserStatusVersionGet(req, db.Redis, db.DB)

	// var resp = &moment.UserVersionGetRep{FriendUID: req.FriendUID, RecommendUID: req.RecommendUID, FollowUID: req.FollowUID}

	// if req.RecommendVersion > req.RecommendVersionRead {
	// 	resp.RecommendState = true
	// }

	// if req.FollowVersion > req.FollowVersionRead {
	// 	resp.FollowState = true
	// }

	// if req.FriendVersion > req.FriendVersionRead {
	// 	resp.FriendState = true
	// }

	return resp, err
}

//UserVersionReadUpdate 获取用户统计忽略配置
func (db *BaseDBService) UserVersionReadUpdate(ctx context.Context, req *moment.UserStatus) (*moment.UserStatus, error) {
	err := cache.UserStatusVersionReadUpdate(req, db.Redis, db.DB)
	return req, err
}

//UserVersionUpdate 获取用户统计忽略配置
func (db *BaseDBService) UserVersionUpdate(ctx context.Context, req *moment.UserVersionUpdateReq) (*moment.UserVersionUpdateRep, error) {
	err := cache.UserStatusVersionUpdate(req, db.Redis, db.DB)
	return &moment.UserVersionUpdateRep{}, err
}

//UserVersionRead 获取用户统计忽略配置
func (db *BaseDBService) UserVersionRead(ctx context.Context, req *moment.UserStatus) (*moment.UserStatus, error) {
	err := cache.UserStatusVersionReadUpdate(req, db.Redis, db.DB)
	return req, err
}

//UserStatusUpdate 更新用户统计(实际上只更新用户朋友圈背景图)
func (db *BaseDBService) UserStatusUpdate(ctx context.Context, req *moment.UserStatus) (*moment.UserStatus, error) {
	var err error
	if req.UserID == 0 {
		return nil, errors.New("UserID不能为空")
	}
	if req.HomeBackground != "" { //更新背景
		if err = internal.UserStatusUpdateDB(&moment.UserStatus{HomeBackground: req.HomeBackground, UserID: req.UserID, UpdateAt: time.Now().Unix()}, db.DB, nil); err != nil {
			return nil, err
		}
	}

	//异步更新(如果存在)用户背景缓存
	go cache.AsyncCacheUpdateUserHomeBackground(db.Redis, int(req.UserID), req.HomeBackground, db.Log)
	// err = cache.UserStatusUpdateCache(req, db.Redis)

	return req, err
}

//
//UserAlbumFromCache 获取用户相册
func (db *BaseDBService) UserAlbumFromCache(ctx context.Context, req *moment.UserAlbumReq) (*moment.UserAlbumRep, error) {
	var resp = &moment.UserAlbumRep{}
	err := cache.UserAlbumCacheGet(req.UserID, resp.Medias, db.Redis, db.DB)
	return resp, err
}

//UserRecommendGet 检查用户是否是推荐用户
func (db *BaseDBService) UserRecommendGet(ctx context.Context, req *moment.UserRecommend) (*moment.UserRecommend, error) {
	var err error
	if req.UserID == 0 {
		return nil, errors.New("UserID参数缺失")
	}
	err = cache.UserRecommendGet(req, db.Redis)
	return req, err
}

//UserAllID  所有开通朋友圈的人
func (db *BaseDBService) UserAllID(ctx context.Context, req *moment.UserAllIDReq) (*moment.UserAllIDRep, error) {
	ids, err := internal.UserAllID(req, db.DB)
	return &moment.UserAllIDRep{UserIDs: ids}, err
}

//UserIgnoreMeGet 获取不让我看的人
func (db *BaseDBService) UserIgnoreMeGet(ctx context.Context, req *moment.UserIgnore) (*moment.UserIgnoreMeGetResp, error) {

	var (
		err  error
		resp = &moment.UserIgnoreMeGetResp{}
	)
	resp.UserIDs, err = internal.UserIgnoreMeGet(req, db.DB)
	return resp, err
}

//UserIgnoresAll 获取不让我看的人
func (db *BaseDBService) UserIgnoresAll(ctx context.Context, req *moment.UserIgnore) (*moment.UserIgnoresAllResp, error) {

	var (
		err  error
		resp = &moment.UserIgnoresAllResp{}
	)
	resp.UserIDs, err = internal.UserIgnoresAllID(req.CreateBy, db.DB)
	return resp, err
}
