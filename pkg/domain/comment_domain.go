package domain

import "gitlab.moments.im/pkg/datamodel"

//CommentDomain 评论领域模型
type CommentDomain struct {
	datamodel.Comment
	Creator *UserDomain `json:"Creator,omitempty"`
	//要回复的用户信息 (如果这条是回复别人的二级回复)
	ReplayUser *UserDomain `json:"ReplayUser,omitempty"`
}
