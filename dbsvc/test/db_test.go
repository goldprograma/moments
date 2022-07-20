package svc

import (
	"context"
	"gitlab.moments.im/dbsvc/svc"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/protoc/moment"
	"reflect"
	"testing"
)

func TestBaseDBService_UserAlbumFromCache(t *testing.T) {

	baseService := &svc.BaseDBService{}

	baseService.RegisterComponent("../config.toml", pkg.Component_DB, pkg.Component_REDIS)

	type args struct {
		ctx context.Context
		req *moment.UserAlbumReq
	}
	tests := []struct {
		name    string
		db      *svc.BaseDBService
		args    args
		want    *moment.UserAlbumRep
		wantErr bool
	}{
		{"111", baseService, args{context.TODO(), &moment.UserAlbumReq{UserID: 1992053}}, nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.db.UserAlbumFromCache(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("BaseDBService.UserAlbumFromCache() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseDBService.UserAlbumFromCache() = %v, want %v", got, tt.want)
			}
		})
	}
}
