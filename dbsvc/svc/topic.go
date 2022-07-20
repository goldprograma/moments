package svc

import (
	"context"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"
	"time"
)

type TopicDBService struct {
	pkg.BaseComponent
}

func (ts *TopicDBService) Add(ctx context.Context, req *moment.Topic) (rep *moment.Topic, err error) {

	req.TopicID = ts.GetID()
	req.CreateAt = time.Now().Unix()
	err = internal.TopicInsert(req, ts.DB)
	return req, err
}

//Delete 删除
func (ts *TopicDBService) Delete(ctx context.Context, req *moment.Topic) (rep *moment.Topic, err error) {

	req.TopicID = ts.GetID()
	req.CreateAt = time.Now().Unix()
	_, err = internal.TopicDelete(req, ts.DB)
	return req, err
}

// Page 获取所有话题
func (ts *TopicDBService) Page(c context.Context, par *moment.TopicPageReq) (*moment.TopicPageRep, error) {
	data, err := internal.TopicPage(par, ts.DB)
	return &moment.TopicPageRep{Topics: data}, err
}

// Types 获取所有话题类型
func (ts *TopicDBService) Types(c context.Context, par *moment.TopicTypeReq) (*moment.TopicTypeRep, error) {
	data, err := internal.TopicTypeAll(ts.DB)
	return &moment.TopicTypeRep{TopicTypes: data}, err
}
