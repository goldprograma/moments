package domain

import "gitlab.moments.im/pkg/datamodel"

//ThumbDomain 点赞领域模型
type ThumbDomain struct {
	datamodel.Thumb
	Creator UserDomain `json:"Creator,omitempty"`
}
