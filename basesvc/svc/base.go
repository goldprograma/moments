package svc

import (
	"context"
	"encoding/json"
	"errors"
	"moments/pkg"
	"moments/pkg/protoc/imapigateway"
	"moments/pkg/protoc/moment"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ViewState_Enable  = 1 // 可见
	ViewState_Disable = 2 //不可见
)

type BaseService struct {
	pkg.BaseComponent
	moment.BaseDBServiceClient
	moment.ForumDBServiceClient
	moment.FollowDBServiceClient
	moment.MediaDBServiceClient
	APIGatewayServiceClient imapigateway.ApiGatewayServiceClient
}

//GetToken 获取token
// @Tags 基础服务
// @Summary 获取token
// @Produce  json
// @Param user-id header int true "user-id"
// @Accept  json
// @Success 200 {object} json.RawMessage
// @Failure 500 {object} json.RawMessage
// @Router /basesvc/token/get [get]
func (bs *BaseService) GetToken(c *gin.Context) {
	var err = errors.New("UserID 不能为空")
	var token string
	userID := c.GetHeader("user-id")
	if userID != "" {
		var userIDInt int64
		if userIDInt, err = strconv.ParseInt(userID, 0, 32); err != nil {
			c.JSON(http.StatusOK, gin.H{
				"State":   true,
				"Code":    "400",
				"Message": "UserID参数类型错误",
				"Data":    "",
			})
			return
		}

		if token, err = pkg.GenerateToken(int32(userIDInt)); err == nil {

			c.JSON(http.StatusOK, gin.H{
				"State":   true,
				"Code":    "200",
				"Message": "获取Token成功",
				"Data":    gin.H{"Token": string(token)},
			})
			return

		}

	}
	c.JSON(http.StatusOK, gin.H{
		"State":   false,
		"Code":    "500",
		"Message": err.Error(),
		"Data":    "",
	})
}

//CheckToken 检查token
// @Tags 基础服务
// @Summary 检查token
// @Produce  json
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/token/check [get]
func (bs *BaseService) CheckToken(c *gin.Context) {
	token := c.Request.Header.Get("authorization")

	result, _ := pkg.CheckJWT(token)
	c.JSON(http.StatusOK, result)
}

//GetUploadDomain 获取上传URL
// @Tags 基础服务
// @Summary 获取上传URL
// @Produce  json
// @Param Location query int true "0为大陆 1为国外"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/uploadurl [get]
func (bs *BaseService) GetUploadDomain(c *gin.Context) {
	location, has := c.GetQuery("Location")
	if !has {
		bs.Response(c, "200", errors.New("参数异常"), "location获取失败", nil)
		return
	}
	if location == "0" {
		bs.Response(c, "200", nil, "获取上传域名成功", bs.Cfg.Uploader.Mainland)
		return
	}
	bs.Response(c, "200", nil, "获取上传域名成功", bs.Cfg.Uploader.Abroad)

}

//TagAdd 新增标签
//ForumAdd 新增标签
// @Tags 标签
// @Summary 新增标签
// @Produce  json
// @Description 只传 Tag.TagName 和 UserTag.UserID
// @Param Authorization header string true "Token"
// @Param body body moment.TagAddReq true "json "
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/tag/add [post]
func (bs *BaseService) TagAdd(c *gin.Context) {
	var params = &moment.TagAddReq{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		bs.Response(c, "TAG_ADD", err, "解析标签参数", nil)
		return
	}
	params.Tag.CreateBy = pkg.GetClaims(c).UserID
	for i := range params.UserTags {
		params.UserTags[i].CreateBy = params.Tag.CreateBy
	}

	if params.Tag.TagName == "" || len(params.UserTags) == 0 {
		bs.Response(c, "TAG_ADD", errors.New("参数检测出错TagName、Users"), "解析标签", nil)
		return
	}
	var data = &moment.TagAddRep{}
	data, err = bs.BaseDBServiceClient.TagAdd(c.Request.Context(), params)
	bs.Response(c, "TAG_ADD", err, "标签", data)
}

//TagGet 查询标签
// @Tags 标签
// @Summary 获取标签
// @Produce  json
// @Accept  json
// @Param Authorization header string true "Token"
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/tag/get [get]
func (bs *BaseService) TagGet(c *gin.Context) {
	var params = &moment.TagGetReq{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		bs.Response(c, "TAG_GET", err, "解析标签参数", nil)
		return
	}
	params.UserID = pkg.GetClaims(c).UserID

	var data = &moment.TagGetRep{}
	data, err = bs.BaseDBServiceClient.TagGet(c.Request.Context(), params)
	bs.Response(c, "TAG_GET", err, "标签获取", data.UserTags)
}

//UserIgnoreAdd 新增标签
//TagAdd 新增标签
//ForumAdd 新增忽略
// @Tags 忽略
// @Summary 新增忽略
// @Produce  json
// @Description 只传 [UserIgnore.UserID、UserIgnore.Look] Look 1不让他看 2不看他的
// @Param body body moment.IgnoreSlice true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/ignore/add [post]
func (bs *BaseService) UserIgnoreAdd(c *gin.Context) {
	var params = &moment.IgnoreSlice{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		bs.Response(c, "USERIGNORE_ADD", err, "解析新增忽略参数", nil)
		return
	}
	var checkResult *imapigateway.CheckIsSelfContactResult
	for i := range params.Ignores {
		if params.Ignores[i].Look > 2 || params.Ignores[i].Look < 1 {
			bs.Response(c, "USERIGNORE_ADD", errors.New("Look值错误"), "Look值错误", nil)
			return
		}
		params.Ignores[i].CreateBy = pkg.GetClaims(c).UserID

		//检查是否是好友,只有好友才可以设置忽略
		if checkResult, err = bs.APIGatewayServiceClient.CheckIsSelfContact(c.Request.Context(), &imapigateway.CheckIsSelfContactReq{SelfId: pkg.GetClaims(c).UserID, UserId: params.Ignores[i].UserID}); err != nil {
			bs.Response(c, "USERIGNORE_ADD", err, "检查是否好友错误", nil)
			return
		}
		if !checkResult.IsSelfContact {
			bs.Response(c, "USERIGNORE_ADD", errors.New("非好友不能忽略"), "非好友不能忽略", nil)
			return
		}

	}
	if len(params.Ignores) == 0 || params.Ignores[0].UserID == 0 || params.Ignores[0].Look == 0 {
		bs.Response(c, "USERIGNORE_ADD", errors.New("参数检测出错UserID、Look"), "解析新增忽略", nil)
		return
	}

	if _, err = bs.BaseDBServiceClient.UserIgnoreAdd(c.Request.Context(), params); err == nil {
		//更新朋友列表里面我能看到的
		go func(params *moment.IgnoreSlice) {
			for _, ignore := range params.Ignores {
				switch ignore.Look {
				case 1: //不让他看
					if _, err = bs.ForumDBServiceClient.ForumFriendDisableView(context.TODO(), &moment.ForumFriend{UserID: ignore.UserID, ViewState: ViewState_Disable, CreateBy: ignore.CreateBy}); err != nil {
						bs.Log.Error("更新朋友帖子可见权限出错", zap.Int32("自己UID", ignore.CreateBy), zap.Int32("更新UID", ignore.UserID), zap.Error(err))
					}
				case 2: //不看他
					if _, err = bs.ForumDBServiceClient.ForumFriendDisableView(context.TODO(), &moment.ForumFriend{UserID: ignore.CreateBy, ViewState: ViewState_Disable, CreateBy: ignore.UserID}); err != nil {
						bs.Log.Error("更新自己帖子可见权限出错", zap.Int32("自己UID", ignore.CreateBy), zap.Int32("更新UID", ignore.UserID), zap.Error(err))
					}
				}
			}
		}(params)
	}

	bs.Response(c, "USERIGNORE_ADD", err, "新增忽略", params.Ignores)
}

//UserIgnoreGet 查询忽略
//TagGet 查询标签
// @Tags 忽略
// @Summary 获取忽略
// @Produce  json
// @Param UserID query int true "朋友ID"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/ignore/get [get]
func (bs *BaseService) UserIgnoreGet(c *gin.Context) {
	var params = &moment.UserIgnore{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		bs.Response(c, "USERIGNORE_GET", err, "解析忽略参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID

	var data = &moment.IgnoreSlice{}
	data, err = bs.BaseDBServiceClient.UserIgnoreGet(c.Request.Context(), params)
	bs.Response(c, "USERIGNORE_GET", err, "忽略获取", data.Ignores)
}

type UserIgnoreCheckResp struct {
	LookMe    bool
	LookOther bool
}

//UserIgnoreCheck 查询忽略
//TagGet 获取用户忽略设置
// @Tags 忽略
// @Summary 获取用户忽略设置
// @Produce  json
// @Param UserID query int true "朋友ID"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/ignore/check [get]
func (bs *BaseService) UserIgnoreCheck(c *gin.Context) {
	var params = &moment.UserIgnore{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		bs.Response(c, "USERIGNORE_CHECK", err, "解析忽略参数", nil)
		return
	}
	params.CreateBy = pkg.GetClaims(c).UserID

	var datas = &moment.IgnoreSlice{}
	datas, err = bs.BaseDBServiceClient.UserIgnoreGet(c.Request.Context(), params)
	var rest = &UserIgnoreCheckResp{}
	for _, ig := range datas.Ignores {
		if ig.Look == 1 {
			rest.LookMe = true
		} else if ig.Look == 2 {
			rest.LookOther = true
		}
	}
	bs.Response(c, "USERIGNORE_CHECK", err, "忽略获取", rest)
}

//UserIgnoreDelete 查询忽略
// @Tags 忽略
// @Summary 删除忽略
// @Produce  json
// @Description 只传 [UserIgnore.UserID、UserIgnore.Look] Look 1不让他看 2不看他的
// @Param body body moment.IgnoreSlice true "json "
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/ignore/delete [post]
func (bs *BaseService) UserIgnoreDelete(c *gin.Context) {
	var params = &moment.IgnoreSlice{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		bs.Response(c, "USERIGNORE_DELETE", err, "解析忽略参数", nil)
		return
	}
	if len(params.Ignores) == 0 {
		bs.Response(c, "USERIGNORE_DELETE", errors.New("数据为空"), "解析标签", nil)
		return
	}

	for i := range params.Ignores {
		params.Ignores[i].CreateBy = pkg.GetClaims(c).UserID
		if params.Ignores[i].Look > 2 || params.Ignores[i].Look < 1 || params.Ignores[i].UserID == 0 {
			bs.Response(c, "USERIGNORE_DELETE", err, "参数错误", nil)
			return
		}

	}
	if _, err = bs.BaseDBServiceClient.UserIgnoreDelete(c.Request.Context(), params); err != nil {
		//更新朋友列表里面我能看到的
		go func(params *moment.IgnoreSlice) {
			for _, ignore := range params.Ignores {
				switch ignore.Look {
				case 1: //不让他看
					if _, err = bs.ForumDBServiceClient.ForumFriendDisableView(context.TODO(), &moment.ForumFriend{UserID: ignore.UserID, ViewState: ViewState_Enable, CreateBy: ignore.CreateBy}); err != nil {
						bs.Log.Error("更新朋友帖子可见权限出错", zap.Int32("自己UID", ignore.CreateBy), zap.Int32("更新UID", ignore.UserID), zap.Error(err))
					}
				case 2: //不看他
					if _, err = bs.ForumDBServiceClient.ForumFriendDisableView(context.TODO(), &moment.ForumFriend{UserID: ignore.CreateBy, ViewState: ViewState_Enable, CreateBy: ignore.UserID}); err != nil {
						bs.Log.Error("更新自己帖子可见权限出错", zap.Int32("自己UID", ignore.CreateBy), zap.Int32("更新UID", ignore.UserID), zap.Error(err))
					}
				}
			}
		}(params)
	}

	bs.Response(c, "USERIGNORE_DELETE", err, "更新忽略", params.Ignores)
}

//UserStatistics 查询用户统计
// @Tags 用户
// @Summary 用户统计
// @Produce  json
// @Param Authorization header string true "Token"
// @Param UserID query int true "朋友ID"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/user/statistics [get]
func (bs *BaseService) UserStatistics(c *gin.Context) {
	var params = &moment.UserStatus{}
	var err error
	if err = c.BindQuery(params); err != nil {
		bs.Response(c, "UserStatus", err, "解析用户参数", nil)
		return
	}
	if params.UserID == 0 {
		params.UserID = pkg.GetClaims(c).UserID
	} else {
		var userInfo *imapigateway.GetUserInfoResult

		if userInfo, err = pkg.GetUserInfoByID(c.Request.Context(), bs.APIGatewayServiceClient, pkg.GetClaims(c).UserID, params.UserID); err != nil {
			bs.Response(c, "UserStatus", err, "查询用户信息出错", nil)
			return
		}
		params.User = userInfo.Uinfo
		//检查我和主人是否是好友
		var checkIsMuteContactResult *imapigateway.CheckIsMuteContactResult
		if checkIsMuteContactResult, err = pkg.CheckIsMuteContact(c.Request.Context(), bs.APIGatewayServiceClient, params.UserID, pkg.GetClaims(c).UserID); err != nil {
			bs.Response(c, "UserStatus", err, "检查是否双向好友失败", err)
			return
		}
		params.IsFriend = checkIsMuteContactResult.Result
	}

	params, err = bs.BaseDBServiceClient.UserStatisticsGet(c.Request.Context(), params)

	bs.Response(c, "UserStatus", err, "查询用户统计", params)
}

//UserHomeBackgroudGet 用户背景图
// @Tags 用户
// @Summary 用户背景图
// @Produce  json
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/user/homebg [get]
func (bs *BaseService) UserHomeBackgroudGet(c *gin.Context) {
	var params = &moment.UserStatus{}
	var err error
	if err = c.BindQuery(params); err != nil {
		bs.Response(c, "UserStatus", err, "解析用户参数", nil)
		return
	}
	if params.UserID == 0 {
		params.UserID = pkg.GetClaims(c).UserID
	}
	params, err = bs.BaseDBServiceClient.UserHomeBackgroudGet(c.Request.Context(), params)
	bs.Response(c, "UserStatus", err, "查询用户背景图", params)
}

//UserVersion 查询用户版本
// @Tags 用户
// @Summary 用户版本
// @Produce  json
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/user/version [get]
func (bs *BaseService) UserVersion(c *gin.Context) {
	var params = &moment.UserStatus{}
	var err error
	if err = c.BindQuery(&params); err != nil {
		bs.Response(c, "USER_VERSION", err, "解析用户统计参数", nil)
		return
	}

	params.UserID = pkg.GetClaims(c).UserID

	var result = &moment.UserVersionGetRep{}
	result, err = bs.BaseDBServiceClient.UserVersionGet(c.Request.Context(), params)
	bs.Response(c, "USER_VERSION", err, "查询用户版本", result)
}

//UserAlbum 用户相册
// @Tags 用户
// @Summary 用户相册
// @Produce  json
// @Param FriendID query int false "朋友ID"
// @Param ForumID query int false "帖子ID"
// @Param Limit query int false "限制条数"
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/user/album [get]
func (bs *BaseService) UserAlbum(c *gin.Context) {
	var params = &moment.ForumGetWithMediaReq{}
	var err error
	if err = c.BindQuery(params); err != nil {
		bs.Response(c, "UserAlbum", err, "解析用户相册参数", nil)
		return
	}
	params.UserID = pkg.GetClaims(c).UserID

	var fs *moment.ForumGetWithMediaResp
	if fs, err = bs.ForumDBServiceClient.ForumGetWithMedia(c.Request.Context(), params); err != nil {
		bs.Response(c, "UserAlbum", err, "解析用户相册", nil)
		return
	}
	var forumsReturn []*moment.ForumFriend

	//查询自己的
	if params.UserID == params.FriendID {
		forumsReturn = fs.Forums
	} else { //查看他人的

		var checkIsMuteContactResult *imapigateway.CheckIsMuteContactResult
		if checkIsMuteContactResult, err = pkg.CheckIsMuteContact(c.Request.Context(), bs.APIGatewayServiceClient, params.UserID, params.FriendID); err != nil {
			bs.Response(c, "UserAlbum", err, "解析用户相册", nil)
			return
		}

		var follow = &moment.Follow{CreateBy: params.UserID, FollowUID: params.FriendID}
		if follow, err = bs.FollowDBServiceClient.Get(c.Request.Context(), follow); err != nil {
			bs.Response(c, "FORUM_OTHER_MAIN", err, "查询粉丝", err)
			return
		}
		for i := range fs.Forums {
			switch fs.Forums[i].Permission {
			case 1:
				forumsReturn = append(forumsReturn, fs.Forums[i])
				continue
			case 3: //粉丝
				if follow.ID > 0 {
					forumsReturn = append(forumsReturn, fs.Forums[i])
					continue
				}
			case 4: //朋友
				if checkIsMuteContactResult.Result {
					forumsReturn = append(forumsReturn, fs.Forums[i])
					continue
				}
			case 5: //给部分人
				if checkIsMuteContactResult.Result {
					var permissionUserIDs []int32
					if err = json.Unmarshal([]byte(fs.Forums[i].PermissionUser), &permissionUserIDs); err != nil {
						bs.Response(c, "FORUM_OTHER_MAIN", err, "解析部分可见人", fs.Forums)
						return
					}
					for _, permissionUserID := range permissionUserIDs {
						if permissionUserID == params.UserID {
							forumsReturn = append(forumsReturn, fs.Forums[i])
							continue
						}
					}
				}
			case 6: //不给
				if checkIsMuteContactResult.Result {
					var permissionUserIDs []int32
					if err = json.Unmarshal([]byte(fs.Forums[i].PermissionUser), &permissionUserIDs); err != nil {
						bs.Response(c, "FORUM_OTHER_MAIN", err, "解析部分不可见人", fs.Forums)
						return
					}
					for _, permissionUserID := range permissionUserIDs {
						if permissionUserID == params.UserID {
							continue
						}
					}
					forumsReturn = append(forumsReturn, fs.Forums[i])
					continue
				}
			}
		}
	}
	medis := make([]*moment.Media, 0)
	for _, forum := range forumsReturn {
		var resp *moment.MediaGetRep
		if resp, err = bs.MediaDBServiceClient.Get(c.Request.Context(), &moment.Media{MainID: forum.ForumID}); err != nil {
			bs.Response(c, "FORUM_OTHER_MAIN", err, "查询帖子媒体出错", fs.Forums)
			return
		}
		medis = append(medis, resp.Medias...)
	}

	bs.Response(c, "UserAlbum", err, "查询用户相册", medis)
}

//UserHomeBackGround 用户朋友圈首页背景
// @Tags 用户
// @Summary 用户朋友圈首页背景设置
// @Produce  json
// @Param body body moment.UserStatus true "json "
// @Description 只传 HomeBackground
// @Param Authorization header string true "Token"
// @Accept  json
// @Success 200 {object} pkg.ResponseMessage
// @Failure 500 {object} pkg.ResponseMessage
// @Router /basesvc/user/homebg [post]
func (bs *BaseService) UserHomeBackgroundUpdate(c *gin.Context) {
	var params = &moment.UserStatus{}
	var err error
	if err = c.BindJSON(&params); err != nil {
		bs.Response(c, "USERHOMEBACKGROUND_ADD", err, "解析用户主页背景参数", nil)
		return
	}
	params.UserID = pkg.GetClaims(c).UserID
	if params.HomeBackground == "" {
		bs.Response(c, "USERHOMEBACKGROUND_ADD", errors.New("HomeBackground为空"), "用户主页背景", nil)
		return
	}
	_, err = bs.BaseDBServiceClient.UserStatusUpdate(c.Request.Context(), params)
	bs.Response(c, "USERHOMEBACKGROUND_ADD", err, "用户主页背景", params)
}
