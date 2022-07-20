package internal

import (
	"encoding/json"
	"fmt"
	"moments/pkg"
	"moments/pkg/models"
	"net/http"

	"github.com/goinggo/mapstructure"
)

//FollowAdd 新增关注
func FollowAdd(cfg *pkg.AppConfig, data []byte) (err error) {

	if _, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/follow/add", "", http.MethodPost, data); err != nil {
		return err
	}

	return err

}

//FollowAddBatch 新增批量关注
func FollowAddBatch(cfg *pkg.AppConfig, follows []*models.Follow) (err error) {
	var data []byte
	data, _ = json.Marshal(&follows)
	_, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/follow/addbatch", "", http.MethodPost, data)

	return err

}

//FollowDeleteBatch 取消关注
func FollowDeleteBatch(cfg *pkg.AppConfig, follows []*models.Follow) (err error) {
	var data []byte
	data, _ = json.Marshal(&follows)
	_, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/follow/deletebatch", "", http.MethodPost, data)

	return err

}

//FollowDelete 取消关注
func FollowDelete(cfg *pkg.AppConfig, data []byte) (err error) {

	_, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/follow/delete", "", http.MethodPost, data)

	return err

}

//FollowGet 取消关注
func FollowGet(cfg *pkg.AppConfig, userID int32, followID int64) (room *models.Follow, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], fmt.Sprintf("/follow/get?UserID=%d&FollowID=%d", userID, followID), "", http.MethodGet, nil); err != nil {
		return nil, err
	}
	room = &models.Follow{}
	if responseMsg.Data != nil {
		if err = mapstructure.Decode(responseMsg.Data, room); err != nil {
			return
		}
	}

	return room, err

}

//FollowPage 查询所有的
func FollowPage(cfg *pkg.AppConfig, createBy int32, followID int64, limit int64) (follows []*models.Follow, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], fmt.Sprintf("/follow/page?CreateBy=%d&FollowID=%d&Limit=%d", createBy, followID, limit), "", http.MethodGet, nil); err != nil {
		return nil, err
	}
	follows = make([]*models.Follow, 0)
	if responseMsg.Data != nil {

		for _, data := range responseMsg.Data.([]interface{}) {
			var one = new(models.Follow)
			if err = mapstructure.Decode(data, one); err != nil {
				return
			}
			follows = append(follows, one)

		}
	}

	return follows, err

}
