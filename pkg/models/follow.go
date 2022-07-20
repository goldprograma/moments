package models

// Follow ... 关注
type Follow struct {
	// ID 主键
	ID int32 `db:"id"`
	// UserID 用户ID
	CreateBy int32 `db:"create_by"`
	FollowID int64 `db:"follow_id"`
	// FollowID 关注人ID
	FollowUID int32 `db:"follow_uid"`
	// CreateAt 关注时间
	CreateAt int64 `db:"create_at"`
}

type FollowPageReq struct {
	CreateBy int32
	FollowID int64
	Limit    int64
}

func (*Follow) TableName() string {
	return " follow "
}
