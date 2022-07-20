package svc

import (
	"context"
	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"
	"time"

	"go.uber.org/zap"
)

type CommentDBService struct {
	pkg.BaseComponent
}

//Delete 删除
func (db *CommentDBService) Delete(ctx context.Context, req *moment.Comment) (rep *moment.Comment, err error) {

	if err = internal.CommentDelete(req, db.DB); err != nil {
		return
	}
	if req.ContentType > 1 {
		err = internal.MediaDelete(&moment.Media{MainID: req.CommentID}, db.DB)
	}

	//删除点赞数据
	var count int64
	if count, err = internal.ThumbCommentDelete(&moment.Thumb{CommentID: req.CommentID}, db.DB); err != nil {
		db.Log.Error("删除评论下所有点赞出错", zap.Int32("用户", req.CreateBy), zap.Int64("评论ID", req.CommentID), zap.Error(err))
		return
	}

	//更新帖子回复数量 评论
	if req.SupID == 0 {
		if err = internal.ForumCommentCountUpdate(req.ForumID, -1, db.DB); err != nil {
			db.Log.Error("更新帖子评论数失败", zap.Int64("帖子ID", req.ForumID), zap.Error(err))
		}
	} else {
		if err = internal.CommentSubCommentCountUpdate(req.SupID, -1, db.DB); err != nil {
			db.Log.Error("更新回复评论数失败", zap.Int64("回复ID", req.SupID), zap.Error(err))
		}
	}

	//更新点赞数据
	err = cache.UserStatusUpdateCache(&moment.UserStatus{UserID: req.CreateBy, ThumbCount: ^count + 1}, db.Redis)
	return req, err
}

//Get 获取单个
func (db *CommentDBService) Get(ctx context.Context, req *moment.Comment) (rep *moment.Comment, err error) {
	err = internal.CommentGet(req, db.DB)
	//查询媒体
	if req.ContentType > 1 {
		req.Medias, err = internal.MediaGet(&moment.Media{MainID: req.CommentID}, db.DB)
	}
	return req, nil
}

//Add 新增帖子
func (db *CommentDBService) Add(ctx context.Context, req *moment.Comment) (rep *moment.Comment, err error) {
	req.CommentID = db.GetID()
	req.CreateAt = time.Now().Unix()
	if req.ContentType > 1 {
		for i := range req.Medias {
			req.Medias[i].MainID = req.CommentID
			req.Medias[i].CreateAt = req.CreateAt
		}
		if err = internal.MediasInsert(req.Medias, db.DB); err != nil {
			return req, err
		}
	}

	if err = internal.CommentInsert(req, db.DB, nil); err != nil {
		return req, err
	}
	//更新帖子回复数量 评论
	if req.SupID == 0 {
		if err = internal.ForumCommentCountUpdate(req.ForumID, 1, db.DB); err != nil {
			db.Log.Error("更新帖子评论数失败", zap.Int64("帖子ID", req.ForumID), zap.Error(err))
		}
	} else {
		if err = internal.CommentSubCommentCountUpdate(req.SupID, 1, db.DB); err != nil {
			db.Log.Error("更新回复评论数失败", zap.Int64("回复ID", req.SupID), zap.Error(err))
		}
	}

	return req, err
}

//Page 查询评论
func (db *CommentDBService) Page(ctx context.Context, req *moment.CommentPageReq) (*moment.CommentPageRep, error) {
	Comments, err := internal.CommentPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	for i := range Comments {
		//查询媒体
		if Comments[i].ContentType > 1 {
			Comments[i].Medias, err = internal.MediaGet(&moment.Media{MainID: Comments[i].CommentID}, db.DB)
		}

		//有多少点赞
		// if Comments[i].ThumbUp, err = internal.ThumbCommentCount(&moment.ThumbPageReq{CommentID: Comments[i].CommentID, CreateBy: Comments[i].CreateBy}, db.DB); err != nil {
		// 	return nil, err
		// }

		//检查是否点赞
		//查点赞ThumbCheck
		if Comments[i].HasThumb, err = internal.ThumbCommentCheck(Comments[i].CommentID, req.CreateBy, db.DB); err != nil {
			return nil, err
		}

		//评论下面有多少回复
		// if Comments[i].SubComments, err = internal.CommentReplayCount(&moment.ReplayPageReq{SupID: Comments[i].CommentID, CreateBy: Comments[i].CreateBy}, db.DB); err != nil {
		// 	return nil, err
		// }
	}
	return &moment.CommentPageRep{Comments: Comments}, nil
}

//PageOrderByThumbup 查询评论
func (db *CommentDBService) PageOrderByThumbup(ctx context.Context, req *moment.CommentPageReq) (*moment.CommentPageRep, error) {
	Comments, err := internal.CommentPageOrderByThumbup(req, db.DB)
	if err != nil {
		return nil, err
	}
	for i := range Comments {
		//查询媒体
		if Comments[i].ContentType > 1 {
			Comments[i].Medias, err = internal.MediaGet(&moment.Media{MainID: Comments[i].CommentID}, db.DB)
		}

		//检查是否点赞
		//查点赞ThumbCheck
		if Comments[i].HasThumb, err = internal.ThumbCommentCheck(Comments[i].CommentID, req.CreateBy, db.DB); err != nil {
			return nil, err
		}

		//评论下面有多少回复
		if Comments[i].SubComments, err = internal.CommentReplayCount(&moment.ReplayPageReq{SupID: Comments[i].CommentID, CreateBy: Comments[i].CreateBy}, db.DB); err != nil {
			return nil, err
		}
	}
	return &moment.CommentPageRep{Comments: Comments}, nil
}

//AllPage 查询评论
func (db *CommentDBService) AllPage(ctx context.Context, req *moment.CommentPageReq) (*moment.CommentPageRep, error) {
	Comments, err := internal.CommentAllPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	for i := range Comments {
		//查询媒体
		if Comments[i].ContentType > 1 {
			Comments[i].Medias, err = internal.MediaGet(&moment.Media{MainID: Comments[i].CommentID}, db.DB)
		}
	}
	return &moment.CommentPageRep{Comments: Comments}, nil
}

//ReplayPage 查询回复
func (db *CommentDBService) ReplayPage(ctx context.Context, req *moment.ReplayPageReq) (*moment.ReplayPageResp, error) {
	Comments, err := internal.ReplayCommentPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	for i := range Comments {
		//查询媒体
		if Comments[i].ContentType > 1 {
			Comments[i].Medias, err = internal.MediaGet(&moment.Media{MainID: Comments[i].CommentID}, db.DB)
		}
		//检查是否点赞
		//查点赞ThumbCheck
		if Comments[i].HasThumb, err = internal.ThumbCommentCheck(Comments[i].CommentID, req.CreateBy, db.DB); err != nil {
			return nil, err
		}

		if Comments[i].ThumbUp, err = internal.ThumbCommentCount(&moment.ThumbPageReq{CommentID: Comments[i].CommentID, Friends: req.Friends, CreateBy: Comments[i].CreateBy}, db.DB); err != nil {
			return nil, err
		}

		//评论下面有多少回复
		// if Comments[i].SubComments, err = internal.CommentFriendCount(&moment.CommentPageReq{SupID: Comments[i].CommentID}, db.DB); err != nil {
		// 	return nil, err
		// }
	}
	return &moment.ReplayPageResp{Comments: Comments}, nil
}

//GetCommentCount 查询评论数量
// func (db *CommentDBService) GetCommentCount(ctx context.Context, req *moment.ReplayPageReq) (*moment.GetCommentCountResp, error) {
// 	var count int64
// 	var err error
// 	if count, err = internal.CommentReplayCount(req, db.DB); err != nil {
// 		return &moment.GetCommentCountResp{}, err
// 	}
// 	return &moment.GetCommentCountResp{Count: count}, nil
// }
