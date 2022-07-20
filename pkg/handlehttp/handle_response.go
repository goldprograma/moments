package handlehttp

import (
	"github.com/gin-gonic/gin"
)

// ResponseMessage 消息返回类型定义
type ResponseMessage struct {
	State   int
	Code    string
	Message string
	Data    interface{} `json:"Data,omitempty"`
}

func handleRespOK(ctx *gin.Context, status int, code string, msg string, data interface{}) {
	ctx.JSON(200, ResponseMessage{
		State:   status,
		Code:    "SUC_" + code,
		Message: msg,
		Data:    data,
	})
}

func handleRespErr(ctx *gin.Context, status int, code string, msg string) {
	ctx.JSON(200, ResponseMessage{
		State:   status,
		Code:    "ERR_" + code,
		Message: msg,
	})
}

// ok
func HandleResp200(ctx *gin.Context, code string, data interface{}) {
	handleRespOK(ctx, 200, code, "请求成功", data)
}

// created
func HandleResp201(ctx *gin.Context, code string, data interface{}) {
	handleRespOK(ctx, 201, code, "操作成功", data)
}

// not modified
func HandleResp304(ctx *gin.Context, code string) {
	handleRespOK(ctx, 304, code, "未更改", nil)
}

// bad request
func HandleResp400(ctx *gin.Context, code string) {
	handleRespErr(ctx, 400, code, "请求参数不正确")
}

// unauthorized
func HandleResp401(ctx *gin.Context, code string) {
	handleRespErr(ctx, 401, code, "无权访问")
}

// forbidden
func HandleResp403(ctx *gin.Context, code string) {
	handleRespErr(ctx, 403, code, "禁止访问")
}

// not found
func HandleResp404(ctx *gin.Context, code string) {
	handleRespErr(ctx, 404, code, "资源不存在")
}

// too much request
func HandleResp429(ctx *gin.Context, code string) {
	handleRespErr(ctx, 429, code, "请求过多")
}

// not found
func HandleResp500(ctx *gin.Context, code string) {
	handleRespErr(ctx, 500, code, "服务器出错")
}
