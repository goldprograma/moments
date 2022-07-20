package models

// Comment ...
type Comment struct {
	// ID 主键
	ID int32 `db:"id"`
	// CommentID 评论id
	CommentID int64 `db:"comment_id"`
	// ForumID 主贴id
	ForumID int64 `db:"forum_id"`
	// Content 内容
	Content string `db:"content"`
	// ReplayID 回复id
	ReplayID int64 `db:"replay_id"`
	// ReplayUID 回复人
	ReplayUID int32 `db:"replay_uid"`
	// ReplayUname 回复人名称
	ReplayUname string `db:"replay_uname"`
	// SupID 一级评论ID
	SupID int64 `db:"sup_id"`
	// SupUser 一级评论人
	SupUser int32 `db:"sup_user"`
	// CreateAt 回复时间
	CreateAt int64 `db:"create_at"`
	// CreateBy 回复人
	CreateBy    int32 `db:"create_by"`
	SubComments int64
	// 媒体文件
	Medias []*Media
}

func (*Comment) TableName() string {
	return " comment "
}

type CommentReq struct {
	ForumID   int64
	SupID     int64
	CommentID int64
	Limit     int64
}
