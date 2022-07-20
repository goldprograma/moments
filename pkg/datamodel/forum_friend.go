package datamodel

// ForumFriend model
type ForumFriend struct {
	ID int64 `db:"id" json:"-"`
	//用户发动态时，会给每个好友的user_id插入一条动态
	UserID  int   `db:"user_id" json:"UserID"`
	ForumID int64 `db:"forum_id" json:"ForumID"`
	//未使用
	Type int `db:"type" json:"-"`
	//动态类型：1 纯文本 2 图片动态 3 视频动态
	ContentType int `db:"content_type" json:"ContentType"`
	//动态文本内容
	Content         string  `db:"content" json:"Content"`
	Longitude       float64 `db:"longitude" json:"Longitude,omitempty"`
	Latitude        float64 `db:"latitude" json:"Latitude,omitempty"`
	LocationCity    string  `db:"location_city" json:"LocationCity,omitempty"`
	LocationName    string  `db:"location_name" json:"LocationName,omitempty"`
	LocationAddress string  `db:"location_address" json:"LocationAddress,omitempty"`
	CreateAt        int     `db:"create_at" json:"CreateAt"`
	//动态作者
	CreateBy      int    `db:"create_by" json:"CreateBy"`
	ContentEntity string `db:"content_entity" json:"ContentEntity,omitempty"`
	//json 整数数组，保存@用户的id
	Mention string `db:"mention" json:"-"`
	//json数组 未使用
	Topic string `db:"topic" json:"-"`
	//动态可见性权限：1公开 2私密 3好友可见 4粉丝可见 5指定用户可见 6指定用户不可见
	Permission int `db:"permission" json:"-"`
	//json整数数组，权限为指定用户可见或不可见时保存用户id
	PermissionUser string `db:"permission_user" json:"-"`
	//点赞数
	ThumbUp int `db:"thumb_up" json:"ThumbUp"`
	//评论数
	CommentCount int `db:"comment_count" json:"CommentCount"`
	//逻辑删除字段 1 正常  0 逻辑删除
	ViewState int `db:"view_state" json:"-"`
}
