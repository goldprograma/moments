package pkg

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var httpClient = &http.Client{
	Timeout: time.Second * 10,
}

type HttpFunc func(string, string, string, string, string) ([]byte, error)

// RequestMessage 客户端请求消息类型定义
type RequestMessage struct {
	Version string
	Data    string
}

// // ResponseMessage 消息返回类型定义
// type ResponseMessage struct {
// 	// 错误码
// 	Code string
// 	// 返回请求成功还是失败
// 	Message string
// 	// 返回数据信息
// 	Data interface{}
// }

func HttpClientFn(domain, path, contentType, method string, postData []byte) (resp *ResponseMessage, err error) {
	var (
		request  *http.Request
		response *http.Response
		body     []byte
	)

	log.Printf("http call url=%s contentType=%s method=%s postData=%s", domain+path, contentType, method, postData)
	switch m := strings.ToUpper(method); m {
	case "POST", "PUT":
		if request, err = http.NewRequest(m, domain+path, bytes.NewBuffer(postData)); err != nil {
			return nil, err
		}
	case "GET", "DELETE":
		if request, err = http.NewRequest(m, domain+path, nil); err != nil {
			return nil, err
		}
	default:
		log.Println("method 错误")
		return nil, errors.New("method 错误")
	}

	request.Header.Set("Content-type", contentType)
	if response, err = httpClient.Do(request); err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if body, err = ioutil.ReadAll(response.Body); err != nil {
		log.Println("Read response.Body error", err)
	}
	if err = json.Unmarshal(body, &resp); err != nil {
		return nil, err
	}

	log.Println("HttpClientFn返回数据", string(body))

	if response.StatusCode != 200 {
		log.Println("status code error=", response.StatusCode)
		return nil, errors.New(resp.Data.(string))
	}

	return resp, err
}
