package pkg

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

var httpCli = &http.Client{
	Timeout: time.Second * 10,
}

//SendRotMsg 机器人发送消息
func sendRotMsg(robotCfg robot, content string) error {
	postJSON := `{
		"chat_room_type": %d,                                          
		"chat_room_id": %d,                                       
		"content_text": "%s"    
	}`
	var path = "/" + robotCfg.Token + "/SendMessage"

	req, err := http.NewRequest(http.MethodPost, robotCfg.Endpoint+path, strings.NewReader(fmt.Sprintf(postJSON, robotCfg.RoomType, robotCfg.RoomID, content)))
	if err != nil {
		return err
	}
	req.Header.Add("content-type", "application/json")
	var resp *http.Response
	if resp, err = httpCli.Do(req); err != nil {
		log.Println("发送错误机器人消息出错", err)
	}
	defer resp.Body.Close()
	return err
}
