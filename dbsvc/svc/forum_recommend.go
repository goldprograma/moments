package svc

import (
	"context"

	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg/protoc/moment"

	"go.uber.org/zap"
)

//GetRecommend 获取单个
func (db *ForumDBService) GetRecommend(ctx context.Context, req *moment.ForumGetReq) (rep *moment.ForumRecommend, err error) {
	rep = &moment.ForumRecommend{ForumID: req.ForumID}

	if err = internal.ForumRecommendGet(rep, db.DB); err != nil {
		return rep, err
	}
	//查询媒体
	if rep.ContentType > 1 {
		rep.Medias, err = internal.MediaGet(&moment.Media{MainID: req.ForumID}, db.DB)
	}
	return rep, err
}

//AddRecommend 新增推荐帖子
func (db *ForumDBService) AddRecommend(ctx context.Context, req *moment.ForumRecommend) (rep *moment.ForumRecommend, err error) {
	if err = internal.ForumRecommendInsert(req, db.DB, nil); err != nil {
		return req, err
	}

	return req, err
}

//RecommendPage 查询话题帖子
func (db *ForumDBService) RecommendPage(ctx context.Context, req *moment.RecommendPageReq) (*moment.RecommendPageRep, error) {
	forums, err := internal.ForumRecommendPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	for i := range forums {
		//查询媒体
		if forums[i].ContentType > 1 {
			forums[i].Medias, err = internal.MediaGet(&moment.Media{MainID: forums[i].ForumID}, db.DB)
		}
		//是否关注

		if forums[i].HasFollow, err = internal.FollowCheck(req.UserID, forums[i].CreateBy, db.DB); err != nil {
			return nil, err
		}

		//查点赞ThumbCheck
		if forums[i].HasThumb, err = internal.ThumbForumCheck(forums[i].ForumID, req.UserID, db.DB); err != nil {
			return nil, err
		}
		//查评论
		if forums[i].Comments, err = internal.CommentPage(&moment.CommentPageReq{ForumID: forums[i].ForumID, Limit: 3}, db.DB); err != nil {
			return nil, err
		}
		//查点赞的人
		if forums[i].ThumbUsers, err = internal.ThumbForumUsers(forums[i].ForumID, 0, 10, db.DB); err != nil {
			return nil, err
		}

		var userRecommend = &moment.UserRecommend{UserID: forums[i].CreateBy}
		//查询回复限制等级
		if err = cache.UserRecommendGet(userRecommend, db.Redis); err != nil {
			return nil, err
		}
		forums[i].LimitVIP = userRecommend.LimitVIP
		comments := forums[i].Comments

		//判断是否都是我和发帖人的对话
		var isBreak bool
		var commentCount = len(comments)
		var lastCommentID int64
		for commentCount > 0 && !isBreak {
			for _, comment := range comments {
				lastCommentID = comment.CommentID
				if !((comment.CreateBy == req.UserID && comment.ReplayUID == forums[i].CreateBy) || (comment.CreateBy == forums[i].CreateBy && comment.ReplayUID == req.UserID)) {
					isBreak = true
					break
				}
				if commentCount > 3 {
					forums[i].Comments = append(forums[i].Comments, comment)
				}
			}

			//查评论
			if comments, err = internal.CommentPage(&moment.CommentPageReq{ForumID: forums[i].ForumID, CommentID: lastCommentID, Limit: 10}, db.DB); err != nil {
				return nil, err
			}

			if len(comments) == 0 {
				break
			}
			commentCount += len(comments)

		}

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
	return &moment.RecommendPageRep{Forums: forums}, nil
}
