package models

// UserRecommend ...
type UserRecommend struct {
	// ID 主键
	ID int32 `json:"id"`
	// ForumType 帖子类型
	ForumType int32 `json:"forum_type"`
	// UserID 用户id
	UserID int32 `json:"user_id"`
	// CreateAt 创建时间
	CreateAt int64 `json:"create_at"`
	// CreateBy 创建人
	CreateBy int32 `json:"create_by"`
}

func (*UserRecommend) TableName() string {
	return " user_recommend "
}
