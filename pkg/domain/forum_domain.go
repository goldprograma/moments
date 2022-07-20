package domain

import "gitlab.moments.im/pkg/datamodel"

//ForumDomain "动态"领域
type ForumDomain struct {
	//动态
	datamodel.ForumFriend
	//逻辑字段：图片与视频媒体文件列表，关联media表
	Medias []datamodel.Media `json:"Medias,omitempty"`
	//逻辑字段：@用户和URL高亮，以json对象数组存入ContentEntity字段
	Entities []ContentEntity `json:"Entitys,omitempty"`
	//逻辑字段：作者信息
	Creator *UserDomain `json:"Creator,omitempty"`
	//逻辑字段: 是否是推荐动态
	IsRecommend bool `json:"IsRecommend"`
	//逻辑字段：评论列表
	Comments []CommentDomain `json:"Comments,omitempty"`
	//逻辑字段：点赞用户列表
	ThumbUserInfos []UserDomain `json:"ThumbUserInfos,omitempty"`
	//逻辑字段：当前用户是否点过赞
	HasThumb bool `json:"HasThumb"`
}

//ContentEntity 当前版本只有@用户会用到
type ContentEntity struct {
	//@的用户user_id
	UserID int `json:"UserID,omitempty"`
	//@的用户名"@xxx"
	UserName   string `json:"UserName,omitempty"`
	AccessHash uint64 ` json:"AccessHash,omitempty"`
	//@用户固定=1
	Type int `json:"Type"`
	//字节长度，后端计算
	Limit int `json:"Limit"`
	//在Content中的开始字节位置，后端计算
	Offset int `json:"Offset"`
	//utf8字符长度，后端计算
	ULimit int `json:"ULimit"`
	//在Content中的开始utf8字符位置，后端计算
	UOffset int `json:"UOffset"`
}
