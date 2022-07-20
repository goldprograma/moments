package internal

import (
	"fmt"
	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/models"
	"net/http"

	"github.com/goinggo/mapstructure"
)

//MediaAdd 新增关注
func MediaAdd(cfg *pkg.AppConfig, data []byte) (err error) {

	if _, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/Media/add", "", http.MethodPost, data); err != nil {
		return err
	}

	return err

}

//MediaDelete 取消关注
func MediaDelete(cfg *pkg.AppConfig, data []byte) (room *models.Media, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/Media/delete", "", http.MethodPost, data); err != nil {
		return nil, err
	}
	room = &models.Media{}
	if responseMsg.Data != nil {
		if err = mapstructure.Decode(responseMsg.Data, room); err != nil {
			return
		}
	}

	return room, err

}

//MediaGet 取消关注
func MediaGet(cfg *pkg.AppConfig, userID int32, MediaID int64) (room *models.Media, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], fmt.Sprintf("/Media/get?UserID=%d&MediaID=%d", userID, MediaID), "", http.MethodGet, nil); err != nil {
		return nil, err
	}
	room = &models.Media{}
	if responseMsg.Data != nil {
		if err = mapstructure.Decode(responseMsg.Data, room); err != nil {
			return
		}
	}

	return room, err

}

//MediaAll 查询所有的
func MediaAll(cfg *pkg.AppConfig, userID int32) (Medias []*models.Media, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], fmt.Sprintf("/Media/all?UserID=%d", userID), "", http.MethodGet, nil); err != nil {
		return nil, err
	}
	Medias = make([]*models.Media, 0)
	if responseMsg.Data != nil {

		for _, data := range responseMsg.Data.([]interface{}) {
			var one = new(models.Media)
			if err = mapstructure.Decode(data, one); err != nil {
				return
			}
			Medias = append(Medias, one)

		}
	}

	return Medias, err

}
