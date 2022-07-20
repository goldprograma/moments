package svc

import (
	"context"

	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"
)

type MediaDBService struct {
	pkg.BaseComponent
}

//Get 获取媒体
func (db *MediaDBService) Get(ctx context.Context, req *moment.Media) (rep *moment.MediaGetRep, err error) {
	medias, err := internal.MediaGet(req, db.DB)
	return &moment.MediaGetRep{Medias: medias}, err
}
