package svc

import (
	"context"
	"errors"
	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"
	"time"
)

type NoticeDBService struct {
	pkg.BaseComponent
}

// Get 获取Notice
func (db *NoticeDBService) Get(ctx context.Context, req *moment.Notice) (rep *moment.Notice, err error) {
	err = internal.NoticeGet(req, db.DB)
	return req, err
}

// Add 新增Notice
func (db *NoticeDBService) Add(ctx context.Context, req *moment.Notice) (rep *moment.Notice, err error) {

	req.NoticeID = db.GetID()
	req.CreateAt = time.Now().Unix()
	err = internal.NoticeInsert(req, db.DB)
	return req, err
}

//Delete 删除
func (db *NoticeDBService) Delete(ctx context.Context, req *moment.Notice) (rep *moment.Notice, err error) {

	req.NoticeID = db.GetID()
	req.CreateAt = time.Now().Unix()
	err = internal.NoticeDelete(req, db.DB, nil)
	return req, err
}

// AddBatch 批量新增Notice
func (db *NoticeDBService) AddBatch(ctx context.Context, req *moment.NoticeAddBatchReq) (rep *moment.None, err error) {
	if len(req.Notices) == 0 {
		return nil, errors.New("无更新")
	}
	for i := range req.Notices {
		req.Notices[i].NoticeID = db.GetID()
		req.Notices[i].CreateAt = time.Now().Unix()
	}

	err = internal.NoticeInsertBatch(req.Notices, db.DB)
	return &moment.None{}, err
}

// Page 获取分页通知
func (db *NoticeDBService) Page(ctx context.Context, req *moment.NoticePageReq) (rep *moment.NoticePageRep, err error) {
	notices, err := internal.NoticePage(req, db.DB)
	return &moment.NoticePageRep{Notices: notices}, err
}
