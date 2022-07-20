package models

// Thumb ...
type Thumb struct {
	// ID 主键
	ID int32 `db:"id"`
	// ThumbID点赞Id
	ThumbID int64 `db:"thumb_id"`
	// ForumID 帖子ID
	ForumID int64 `db:"forum_id"`
	// ForumUID帖子拥有者Id
	ForumUID int32 `db:"forum_uid"`
	// UpDown 1点赞2踩
	UpDown int32 `db:"up_down"`
	// CreateAt 创建时间
	CreateAt int64 `db:"create_at"`
	// CreateBy 创建人
	CreateBy int32 `db:"create_by"`
}

func (*Thumb) TableName() string {
	return " thumb "
}

type ThumbPageReq struct {
	ThumbID int64
	ForumID int64
	UpDown  int32
	Limit   int64
}
