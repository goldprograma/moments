package datamodel

// Thumb model
type Thumb struct {
	ID      int   `db:"id" json:"-"`
	ThumbID int64 `db:"thumb_id" json:"ThumbID,omitempty"`
	//点赞的动态
	ForumID int64 `db:"forum_id" json:"ForumID,omitempty"`
	//点赞动态的create_by
	ForumUID int `db:"forum_uid" json:"ForumUID,omitempty"`
	//废弃：不能点赞回复
	CommentID int64 `db:"comment_id" json:"CommentID,omitempty"`
	//废弃：不能点赞回复
	CommentUID int `db:"comment_uid" json:"CommentUID,omitempty"`
	//只能点赞，固定=1
	UpDown   int `db:"up_down" json:"UpDown,omitempty"`
	CreateAt int `db:"create_at" json:"CreateAt,omitempty"`
	//点赞用户
	CreateBy int `db:"create_by" json:"CreateBy,omitempty"`
}
