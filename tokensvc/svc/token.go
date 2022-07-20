package svc

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"gitlab.moments.im/pkg"
	"gitlab.moments.im/pkg/handlehttp"
	"gitlab.moments.im/pkg/protoc/imapigateway"
	"gitlab.moments.im/pkg/protoc/moment"
	"go.uber.org/zap"

	"github.com/gin-gonic/gin"
)

type ITokenService interface {
	Routers(app *gin.Engine)
	GetToken(c *gin.Context)
	CheckToken(c *gin.Context)
	UserToken(ctx context.Context, req *moment.UserTokenReq) (*moment.UserTokenRep, error)
}

type TokenService struct {
	// pkg.BaseComponent
	imGrpcClient imapigateway.ApiGatewayServiceClient
	logger       *zap.Logger
}

func NewTokenService(imGrpcClient imapigateway.ApiGatewayServiceClient, logger *zap.Logger) ITokenService {
	return &TokenService{imGrpcClient: imGrpcClient, logger: logger}
}

//GetToken 获取token
// @Tags Token服务
// @Summary 获取token
// @Produce  json
// @Param user-id header int true "user-id"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /tokensvc/get [get]
func (bs *TokenService) GetToken(ctx *gin.Context) {
	// var err = errors.New("UserID 不能为空")

	user_id_str := ctx.GetHeader("user-id")
	if user_id_str == "" {
		handlehttp.HandleResp400(ctx, "USER_ID_IS_REQUIRED")
		return
	}
	user_id, err := strconv.Atoi(user_id_str)
	if err != nil {
		bs.logger.Error("TokenService GetToken => " + err.Error())
		handlehttp.HandleResp400(ctx, "USER_ID_INCORRECT")
		return
	}
	isok, err := validateUserInfo(bs.imGrpcClient, user_id)
	if err != nil {
		bs.logger.Error("TokenService validateUserInfo => " + err.Error())
		handlehttp.HandleResp500(ctx, "VALIDATE_USER_ID_ERROR")
		return
	}
	if !isok {
		handlehttp.HandleResp400(ctx, "USER_ID_INCORRECT")
		return
	}
	token, err := pkg.GenerateToken(int32(user_id))
	if err != nil {
		bs.logger.Error("TokenService GenerateToken => " + err.Error())
		handlehttp.HandleResp500(ctx, "GENERATE_TOKEN")
		return
	}
	handlehttp.HandleResp200(ctx, "GENERATE_TOKEN", gin.H{
		"Token": token,
	})

	// if userID != "" {
	// 	var userIDInt int64
	// 	if userIDInt, err = strconv.ParseInt(userID, 0, 32); err != nil {
	// 		c.JSON(http.StatusOK, gin.H{
	// 			"State":   200,
	// 			"Code":    "400",
	// 			"Message": "UserID参数类型错误",
	// 			"Data":    "",
	// 		})
	// 		return
	// 	}

	// 	if token, err = pkg.GenerateToken(int32(userIDInt)); err == nil {

	// 		c.JSON(http.StatusOK, gin.H{
	// 			"State":   200,
	// 			"Code":    "200",
	// 			"Message": "获取Token成功",
	// 			"Data":    gin.H{"Token": string(token)},
	// 		})
	// 		return

	// 	}

	// }
	// c.JSON(http.StatusOK, gin.H{
	// 	"State":   400,
	// 	"Code":    "500",
	// 	"Message": err.Error(),
	// 	"Data":    "",
	// })
}

//CheckToken 检查token
// @Tags Token服务
// @Summary 检查token
// @Produce  json
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /tokensvc/check [get]
func (bs *TokenService) CheckToken(c *gin.Context) {
	token := c.Request.Header.Get("authorization")

	result, _ := pkg.CheckJWT(token)
	c.JSON(http.StatusOK, result)
}

//UserToken 获取用户token
func (bs *TokenService) UserToken(ctx context.Context, req *moment.UserTokenReq) (reply *moment.UserTokenRep, err error) {
	if req.UserID <= 0 {
		return reply, errors.New("user_id incorrect")
	}

	isok, err := validateUserInfo(bs.imGrpcClient, int(req.UserID))
	if err != nil {
		return reply, err
	}
	if !isok {
		return reply, errors.New("user_id incorrect")
	}

	token, err := pkg.GenerateToken(req.UserID)
	if err != nil {
		return reply, errors.New("generate token failed")
	}

	return &moment.UserTokenRep{Token: token, Code: 1}, nil
}

/* -------- */

//通过im校验user_id
func validateUserInfo(imGrpcClient imapigateway.ApiGatewayServiceClient, user_id int) (bool, error) {
	req := &imapigateway.GetUserInfoByIdReq{
		SelfId: int32(user_id),
		UserId: int32(user_id),
	}
	res, err := imGrpcClient.GetUserInfoById(context.TODO(), req)
	return res != nil, err
}
