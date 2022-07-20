package datamodel

// UserRecommend model
type UserRecommend struct {
	ID        int    `db:"id" json:"ID"`
	ForumType int    `db:"forum_type" json:"ForumType,omitempty"`
	UserID    int    `db:"user_id" json:"UserID"`
	CreateAt  int    `db:"create_at" json:"CreateAt"`
	CreateBy  int    `db:"create_by" json:"CreateBy"`
	Mark      string `db:"mark" json:"Mark"`
	LimitVip  int    `db:"limit_vip" json:"LimitVip"`
}
