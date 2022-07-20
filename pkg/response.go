package pkg

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// ResponseMessage 消息返回类型定义
type ResponseMessage struct {
	State   uint16
	Code    string
	Message string
	Data    interface{}
}

//Response 返回json 数据 data
func (bs *BaseComponent) Response(c *gin.Context, code string, err error, msg string, data interface{}) {
	if err != nil {
		bs.Log.Error(fmt.Sprintf("%s  err:%v", msg, err))
		c.JSON(200, ResponseMessage{State: 500, Code: "ERR_" + code, Message: msg + "失败", Data: err})
	} else {
		c.JSON(200, ResponseMessage{State: 200, Code: "SUC_" + code, Message: msg + "成功", Data: data})
	}
}
