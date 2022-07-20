package svc

import (
	"context"
	"errors"
	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"
	"time"

	"go.uber.org/zap"
)

type ThumbDBService struct {
	pkg.BaseComponent
}

// Add 新增Thumb
func (db *ThumbDBService) Add(ctx context.Context, req *moment.Thumb) (rep *moment.Thumb, err error) {
	// 帖子号
	if (req.ForumID == 0 && req.CommentID == 0) || req.CreateBy == 0 {
		return req, errors.New("没传ForumID、CreateBy、CommentID")
	}
	req.ThumbID = db.GetID()
	req.CreateAt = time.Now().Unix()
	if err = internal.ThumbInsert(req, db.DB, nil); err != nil {
		return req, err
	}
	if req.CommentID == 0 { //发帖
		//更新点赞数
		if err = internal.ForumThumbUpdate(req.ForumID, 1, db.DB); err != nil {
			db.Log.Error("更新帖子点赞数失败", zap.Int64("帖子", req.ForumID), zap.Error(err))
			return req, nil
		}
	} else { //点赞
		//更新点赞数
		if err = internal.CommentThumbUpdate(req.CommentID, 1, db.DB); err != nil {
			db.Log.Error("更新评论点赞数失败", zap.Int64("评论ID", req.CommentID), zap.Error(err))
			return req, nil
		}
	}
	var userID = req.ForumUID
	if req.CommentID > 0 {
		userID = req.CommentUID
	}
	var userStatus = &moment.UserStatus{UserID: userID, ThumbCount: 1}
	//更新用户点赞数
	if err = internal.UserStatusThumbCountUpdateDB(userStatus, db.DB); err != nil {
		db.Log.Error("更新用户DB点赞数失败", zap.Int32("用户", userID), zap.Error(err))
	}

	//更新用户点赞数
	if err = cache.UserStatusUpdateCache(userStatus, db.Redis); err != nil {
		db.Log.Error("更新用户缓存点赞数失败", zap.Int32("用户", userID), zap.Error(err))
	}

	return req, err
}

// Delete 删除Thumb
func (db *ThumbDBService) Delete(ctx context.Context, params *moment.Thumb) (rep *moment.Thumb, err error) {
	if params.CreateBy == 0 || (params.ForumID == 0 && params.CommentID == 0) {
		return params, errors.New("参数缺失CreateBy、ForumID、CommentID")
	}

	var count int64

	if params.CommentID == 0 { //发帖
		if count, err = internal.ThumbForumUserDelete(params, db.DB); err != nil {
			return nil, err
		}
		if count > 0 {
			err = internal.ForumThumbUpdate(params.ForumID, ^count+1, db.DB)
		}

	} else { //回复
		if count, err = internal.ThumbCommentUserDelete(params, db.DB); err != nil {
			return nil, err
		}
		if count > 0 {
			err = internal.CommentThumbUpdate(params.CommentID, ^count+1, db.DB)
		}

	}

	var userID = params.ForumUID
	if params.CommentUID > 0 {
		userID = params.CommentUID
	}

	var userStatus = &moment.UserStatus{UserID: userID, ThumbCount: ^count + 1}
	//更新用户点赞数
	if err = internal.UserStatusThumbCountUpdateDB(userStatus, db.DB); err != nil {
		db.Log.Error("更新用户DB点赞数失败", zap.Int32("用户", userID), zap.Error(err))
	}

	//更新用户点赞数
	if err = cache.UserStatusUpdateCache(userStatus, db.Redis); err != nil {
		db.Log.Error("更新用户点赞数失败", zap.Int32("用户", params.CreateBy), zap.Error(err))
	}

	return params, err
}

//HasClick 检查用户是否点赞
func (db *ThumbDBService) HasClick(ctx context.Context, params *moment.Thumb) (rep *moment.HasClickResp, err error) {

	if (params.ForumID == 0 && params.CommentID == 0) || params.CreateBy == 0 {
		return nil, errors.New("ForumID、CommentID、CreateBy缺失")
	}
	var has bool
	if params.CommentID > 0 {
		has, err = internal.ThumbCommentCheck(params.CommentID, params.CreateBy, db.DB)
	} else {
		has, err = internal.ThumbForumCheck(params.ForumID, params.CreateBy, db.DB)
	}

	return &moment.HasClickResp{State: has}, err
}

//UserCount 用户获取点赞数量
func (db *ThumbDBService) UserCount(ctx context.Context, params *moment.ThumbUserCountReq) (rep *moment.ThumbUserCountRep, err error) {

	if params.UserID == 0 {
		return nil, errors.New("UserID缺失")
	}
	var count int64
	count, err = internal.ThumbUserCount(params, db.DB)

	return &moment.ThumbUserCountRep{Count: count}, err
}

// Page 获取点赞
func (db *ThumbDBService) Page(ctx context.Context, params *moment.ThumbPageReq) (rep *moment.ThumbPageRep, err error) {
	var thumbs []*moment.Thumb
	thumbs, err = internal.ThumbPage(params, db.DB)
	return &moment.ThumbPageRep{Thumbs: thumbs}, err
}

// UserID 获取点赞
func (db *ThumbDBService) UserID(ctx context.Context, params *moment.ThumbPageReq) (rep *moment.ThumbUserIDResp, err error) {
	var users []int32
	users, err = internal.ThumbForumUsers(params.ForumID, params.ThumbID, params.Limit, db.DB)
	return &moment.ThumbUserIDResp{UserID: users}, err
}

// GetForumCount 获取点赞数量
func (db *ThumbDBService) GetForumCount(ctx context.Context, params *moment.ThumbPageReq) (rep *moment.GetThumbCountResp, err error) {
	var count int64
	count, err = internal.ThumbForumCount(params, db.DB)
	return &moment.GetThumbCountResp{Count: count}, err
}

// GetCommentCount 获取点赞数量
func (db *ThumbDBService) GetCommentCount(ctx context.Context, params *moment.ThumbPageReq) (rep *moment.GetThumbCountResp, err error) {
	var count int64
	count, err = internal.ThumbCommentCount(params, db.DB)
	return &moment.GetThumbCountResp{Count: count}, err
}
