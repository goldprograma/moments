package datamodel

// Comment model
type Comment struct {
	ID        int `db:"id" json:"-"`
	CommentID int `db:"comment_id" json:"CommentID"`
	//评论相关的动态forum_id
	ForumID int `db:"forum_id" json:"ForumID"`
	//评论相关的动态create_by
	ForumUser int `db:"forum_user" json:"ForumUser"`
	//评论内容
	Content string `db:"content" json:"Content"`
	//评论类型：1=文字，只有文字类型
	ContentType int `db:"content_type" json:"ContentType"`
	//废弃：评论中只有纯文本，不需要@用户和URL
	AtEntity string `db:"at_entity" json:"-"`
	//废弃：评论不能被点赞
	ThumbUp int `db:"thumb_up" json:"-"`
	//回复哪一条回复
	ReplayID int `db:"replay_id" json:"ReplayID,omitempty"`
	//回复哪个用户
	ReplayUID int `db:"replay_uid" json:"ReplayUID,omitempty"`
	//无用：等于ReplayID即可
	SupID int `db:"sup_id" json:"-"`
	//无用：等于ReplayUID即可
	SupUser  int `db:"sup_user" json:"-"`
	CreateAt int `db:"create_at" json:"CreateAt,omitempty"`
	CreateBy int `db:"create_by" json:"CreateBy,omitempty"`
	//废弃：不需要显示评论自己的回复数
	SubComments int `db:"sub_comments" json:"-"`
}
