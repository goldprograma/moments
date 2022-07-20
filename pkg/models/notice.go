package models

import "gitlab.moments.im/pkg"

// Notice ...
type Notice struct {
	// ID 主键
	ID int32 `db:"id"`
	// NoticeID 通知消息ID
	NoticeID int64 `db:"notice_id"`
	// Type 1点赞 2评论
	Type pkg.NoticeType `db:"type"`
	// RelationID 帖子id
	RelationID int64 `db:"relation_id"`
	// Notifier 通知人
	Notifier int32 `db:"notifier"`
	// Status 1为已读
	Status int8 `db:"status"`
	// CreateAt 通知时间
	CreateAt int64 `db:"create_at"`
	// CreateBy 通知创建人
	CreateBy int32 `db:"create_by"`
	Content  NoticeContent
}

type NoticeContent struct {
	CommentID      int64  `db:"comment_id"`
	CommentContent string `db:"comment_content"`
	SupID          int64  `db:"sup_id"`
	SupContent     string `db:"sup_content"`
	ForumID        int64  `db:"forum_id"`
	ForumContent   string `db:"forum_content"`
	MediaContent   string `db:"media_content"`
}
type NoticePageReq struct {
	Notifier int32
	NoticeID int64
	Limit    int64
}

func (*Notice) TableName() string {
	return " notice "
}
