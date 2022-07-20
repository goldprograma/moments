package svc

import (
	"context"
	"encoding/json"

	"gitlab.moments.im/dbsvc/cache"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg/protoc/moment"
)

//GetTopic 获取单个
func (db *ForumDBService) GetTopic(ctx context.Context, req *moment.ForumGetReq) (rep *moment.ForumTopic, err error) {
	rep = &moment.ForumTopic{ForumID: req.ForumID}

	if err = internal.ForumTopicGet(rep, db.DB); err != nil {
		return rep, err
	}
	//查询媒体
	if rep.ContentType > 1 {
		rep.Medias, err = internal.MediaGet(&moment.Media{MainID: req.ForumID}, db.DB)
	}
	return rep, nil
}

//AddTopic 新增帖子
func (db *ForumDBService) AddTopic(ctx context.Context, req *moment.ForumTopic) (rep *moment.ForumTopic, err error) {

	tx, err := db.DB.Beginx()
	if err != nil {
		return nil, err
	}
	var topics = make([]*moment.Topic, 0, 3)

	if err = json.Unmarshal([]byte(req.Topic), &topics); err != nil {
		return nil, err
	}

	for _, topic := range topics {
		req.TopicID = topic.TopicID
		if err = internal.ForumTopicInsert(req, db.DB, tx); err != nil {
			return req, err
		}
	}

	tx.Commit()

	return req, err
}

//TopicPage 查询话题帖子
func (db *ForumDBService) TopicPage(ctx context.Context, req *moment.ForumTopicPageReq) (*moment.ForumTopicPageRep, error) {
	forums, err := internal.ForumsTopicPage(req, db.DB)
	if err != nil {
		return nil, err
	}
	var userRecommend = &moment.UserRecommend{}

	for i := range forums {
		//查询VIP
		//查询回复限制等级
		userRecommend.UserID = forums[i].CreateBy
		if err = cache.UserRecommendGet(userRecommend, db.Redis); err != nil {
			return nil, err
		}
		forums[i].IsRecommend = userRecommend.ID > 0
		forums[i].LimitVIP = userRecommend.LimitVIP

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
		//查评论 先评论得在前面
		if forums[i].Comments, err = internal.CommentPage(&moment.CommentPageReq{ForumID: forums[i].ForumID, Limit: 0}, db.DB); err != nil {
			return nil, err
		}

		//查点赞数量
		// if forums[i].ThumbUp, err = internal.ThumbForumCount(&moment.ThumbPageReq{ForumID: forums[i].ForumID}, db.DB); err != nil {
		// 	return nil, err
		// }
		//查点赞的人
		if forums[i].ThumbUsers, err = internal.ThumbForumUsers(forums[i].ForumID, 0, 10, db.DB); err != nil {
			return nil, err
		}
	}
	return &moment.ForumTopicPageRep{Forums: forums}, nil
}
