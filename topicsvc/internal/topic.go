package internal

import (
	"fmt"
	"moments/pkg"
	"moments/pkg/models"
	"net/http"

	"github.com/goinggo/mapstructure"
)

// //TopicAdd 新增关注
// func TopicAdd(cfg *pkg.AppConfig, data []byte) (err error) {

// 	if _, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/topic/add", "", http.MethodPost, data); err != nil {
// 		return err
// 	}

// 	return err

// }

// //TopicDelete 取消关注
// func TopicDelete(cfg *pkg.AppConfig, data []byte) (room *models.Topic, err error) {

// 	var responseMsg *pkg.ResponseMessage
// 	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/topic/delete", "", http.MethodPost, data); err != nil {
// 		return nil, err
// 	}
// 	room = &models.Topic{}
// 	if responseMsg.Data != nil {
// 		if err = mapstructure.Decode(responseMsg.Data, room); err != nil {
// 			return
// 		}
// 	}

// 	return room, err

// }

//TopicPage 分页评论
func TopicPage(cfg *pkg.AppConfig, req *models.TopicPageReq) (Topics []*models.Topic, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], fmt.Sprintf("/topic/page?TopicID=%d&TopicTypeID=%d&CreateBy=%d&TopicName=%s&Limit=%d", req.TopicID, req.TopicTypeID, req.CreateBy, req.TopicName, req.Limit), "", http.MethodGet, nil); err != nil {
		return nil, err
	}
	Topics = make([]*models.Topic, 0)
	if responseMsg.Data != nil {
		for _, data := range responseMsg.Data.([]interface{}) {
			var one = new(models.Topic)
			if err = mapstructure.Decode(data, one); err != nil {
				return
			}
			Topics = append(Topics, one)

		}
	}

	return Topics, err

}

//TopicTypeAll 分页评论
func TopicTypeAll(cfg *pkg.AppConfig) (Topics []*models.TopicType, err error) {

	var responseMsg *pkg.ResponseMessage
	if responseMsg, err = pkg.HttpClientFn(cfg.Depends["dbService"], "/topictype/all", "", http.MethodGet, nil); err != nil {
		return nil, err
	}
	Topics = make([]*models.TopicType, 0)
	if responseMsg.Data != nil {
		for _, data := range responseMsg.Data.([]interface{}) {
			var one = new(models.TopicType)
			if err = mapstructure.Decode(data, one); err != nil {
				return
			}
			Topics = append(Topics, one)

		}
	}

	return Topics, err

}
