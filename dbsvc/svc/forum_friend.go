package svc

import (
	"context"
	"time"

	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"

	"go.uber.org/zap"
)

type ForumDBService struct {
	pkg.BaseComponent
}

//Delete 删除
func (db *ForumDBService) Delete(ctx context.Context, req *moment.ForumFriend) (rep *moment.ForumFriend, err error) {
	rep = &moment.ForumFriend{}
	var count int64
	if count, err = internal.ForumDelete(req, db.DB); err != nil {
		return
	}
	if count > 0 {
		go func(req *moment.ForumFriend, db *ForumDBService) {

			if err = internal.MediaDelete(&moment.Media{MainID: req.ForumID}, db.DB); err != nil {
				db.Log.Error("删除媒体出错", zap.Error(err))
			}

			if err = internal.CommentDelete(&moment.Comment{ForumID: req.ForumID}, db.DB); err != nil {
				db.Log.Error("删除评论出错", zap.Error(err))

			}

			if _, err = internal.ThumbForumDelete(&moment.Thumb{ForumID: req.ForumID}, db.DB); err != nil {
				db.Log.Error("删除点赞出错", zap.Error(err))

			}

			if err = internal.UserStatusForumCountUpdateDB(&moment.UserStatus{UserID: req.CreateBy, ForumCount: -1}, db.DB); err != nil {
				db.Log.Error("更新用户发帖数出错", zap.Error(err))
			}

			//更新自己的发帖数
			if err = cache.UserStatusUpdateCache(&moment.UserStatus{UserID: req.CreateBy, ForumCount: -1}, db.Redis); err != nil {
				db.Log.Error("删除帖子后更新用户发帖出错", zap.Int32("用户", req.CreateBy), zap.Int64("帖子", req.ForumID), zap.Error(err))
			}
		}(req, db)
	}

	return
}

//UserDelete 删除
func (db *ForumDBService) UserDelete(ctx context.Context, req *moment.ForumFriend) (rep *moment.ForumFriend, err error) {
	err = internal.ForumUserDelete(req, db.DB)
	return
}

//Ignore 设置忽略帖子
func (db *ForumDBService) Ignore(ctx context.Context, req *moment.ForumIgnore) (rep *moment.ForumIgnore, err error) {
	//删除朋友
	if err = internal.ForumFriendDisableView(&moment.ForumFriend{UserID: req.CreateBy, ForumID: req.ForumID}, db.DB); err != nil {
		return req, err
	}
	req.IgnoreID = db.GetID()
	err = internal.IgnoreForum(req, db.DB)

	return req, err
}

//Check 检查
func (db *ForumDBService) Check(ctx context.Context, req *moment.ForumFriend) (rep *moment.ForumFriend, err error) {
	err = internal.ForumFriendGet(req, db.DB)
	return req, nil
}

//Get 获取单个
func (db *ForumDBService) Get(ctx context.Context, req *moment.ForumGetReq) (rep *moment.ForumFriend, err error) {
	rep = &moment.ForumFriend{ForumID: req.ForumID}

	if err = internal.ForumFriendGet(rep, db.DB); err != nil {
		return rep, err
	}
	//查询媒体
	if rep.ContentType > 1 {
		rep.Medias, err = internal.MediaGet(&moment.Media{MainID: req.ForumID}, db.DB)
	}
	return rep, nil
}

//AddMySelf 新增帖子
func (db *ForumDBService) AddMySelf(ctx context.Context, req *moment.ForumFriend) (rep *moment.ForumFriend, err error) {
	req.ForumID = db.GetID()
	req.CreateAt = time.Now().Unix()
	if req.ContentType > 1 {
		for i := range req.Medias {
			req.Medias[i].MainID = req.ForumID
			req.Medias[i].CreateAt = req.CreateAt
		}
		if err = internal.MediasInsert(req.Medias, db.DB); err != nil {
			return req, err
		}
	}
	if err = internal.ForumFriendInsert(req, db.DB, nil); err != nil {
		return req, err
	}
	userStatus := &moment.UserStatus{UserID: req.CreateBy, ForumCount: 1}

	if err = internal.UserStatusForumCountUpdateDB(userStatus, db.DB); err != nil {
		db.Log.Error("更新自己发帖数出错", zap.Int32("用户", userStatus.UserID), zap.Error(err))
	}
	if req.Permission != 2 {
		if err = cache.UserForumCacheSet(req, db.Redis); err != nil {
			db.Log.Error("缓存自己的帖子出错", zap.Error(err))
		}
	}
	if req.Permission == 1 { //公开 存入缓存相册

	}
	//更新用户统计
	if err = cache.UserStatusUpdateCache(userStatus, db.Redis); err != nil {
		db.Log.Error("更新用户发帖数失败", zap.Error(err))
	}
	return req, err
}

//AddFriend 新增帖子
func (db *ForumDBService) AddFriend(ctx context.Context, req *moment.ForumFriend) (rep *moment.ForumFriend, err error) {
	err = internal.ForumFriendInsert(req, db.DB, nil)
	return req, err
}

func (db *ForumDBService) GetNewestFromCache(ctx context.Context, req *moment.ForumFriend) (*moment.ForumFriend, error) {
	err := cache.UserForumCacheGet(req, db.Redis, db.DB)
	return req, err
}

//FollowFourmPage 查询话题帖子

//FriendPage 查询朋友帖子
func (db *ForumDBService) FriendPage(ctx context.Context, req *moment.ForumFriendPageReq) (rep *moment.ForumFriendPageRep, err error) {
	var forums []*moment.ForumFriend
	if forums, err = internal.ForumFriendsPage(req, db.DB); err != nil {
		return rep, err
	}

	if len(forums) > 0 {
		if err = cache.UserStatusVersionReadUpdate(&moment.UserStatus{FollowVersionRead: forums[0].ForumID, UserID: req.UserID}, db.Redis, db.DB); err != nil {
			db.Log.Error("更新用户已读关注版本号出错", zap.Int32("用户", req.UserID), zap.Error(err))
		}
	} else {
		if err = cache.UserStatusVersionReadUpdate(&moment.UserStatus{FollowVersionRead: 0, UserID: req.UserID}, db.Redis, db.DB); err != nil {
			db.Log.Error("更新用户已读关注版本号出错", zap.Int32("用户", req.UserID), zap.Error(err))
		}
	}
	return &moment.ForumFriendPageRep{Forums: forums}, nil
}

//ForumGetWithMedia 查询帖子带媒体的
func (db *ForumDBService) ForumGetWithMedia(ctx context.Context, req *moment.ForumGetWithMediaReq) (rep *moment.ForumGetWithMediaResp, err error) {
	var forums []*moment.ForumFriend
	forums, err = internal.ForumGetWithMedia(req, db.DB)
	return &moment.ForumGetWithMediaResp{Forums: forums}, err
}

//SelfMainPage 查询自己的帖子
func (db *ForumDBService) SelfMainPage(ctx context.Context, req *moment.SelfMainPageReq) (*moment.SelfMainPageRep, error) {
	forums, err := internal.ForumSelfMainPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	for i := range forums {
		//查询媒体
		if forums[i].ContentType > 1 {
			if forums[i].Medias, err = internal.MediaGet(&moment.Media{MainID: forums[i].ForumID}, db.DB); err != nil {
				return nil, err
			}
		}
		//是否关注

		//查点赞ThumbCheck
		if forums[i].HasThumb, err = internal.ThumbForumCheck(forums[i].ForumID, req.UserID, db.DB); err != nil {
			return nil, err
		}
		//查评论 先评论得在前面
		if forums[i].Comments, err = internal.CommentPage(&moment.CommentPageReq{ForumID: forums[i].ForumID, Limit: 0}, db.DB); err != nil {
			return nil, err
		}

		//查点赞的人
		if forums[i].ThumbUsers, err = internal.ThumbForumUsers(forums[i].ForumID, 0, 10, db.DB); err != nil {
			return nil, err
		}
	}
	return &moment.SelfMainPageRep{Forums: forums}, nil
}

//OtherMainPage 查询其他人的帖子
func (db *ForumDBService) OtherMainPage(ctx context.Context, req *moment.OtherMainPageReq) (*moment.OtherMainPageRep, error) {
	forums, err := internal.ForumOtherMainPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	return &moment.OtherMainPageRep{Forums: forums}, nil
}

//ForumOtherMainByMouth 查询其他人的帖子
func (db *ForumDBService) ForumOtherMainByMouth(ctx context.Context, req *moment.ForumOtherMainByMouthReq) (*moment.ForumOtherMainByMouthResp, error) {
	forums, err := internal.ForumOtherMainByMouth(req, db.DB)
	if err != nil {
		return nil, err
	}
	return &moment.ForumOtherMainByMouthResp{Forums: forums}, nil
}

//ParticipatingFriends 查询帖子参与用户不包含帖子主
func (db *ForumDBService) ParticipatingFriends(ctx context.Context, req *moment.ParticipatingFriendsMsg) (*moment.ParticipatingFriendsMsg, error) {
	user, err := internal.ParticipatingFriends(req, db.DB)
	return &moment.ParticipatingFriendsMsg{Friends: user}, err
}

//ForumFriendDisableView 更新帖子观看权限
func (db *ForumDBService) ForumFriendDisableView(ctx context.Context, req *moment.ForumFriend) (*moment.ForumFriend, error) {
	err := internal.ForumFriendDisableView(req, db.DB)
	return req, err
}
