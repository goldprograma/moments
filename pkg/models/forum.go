package models

import "gitlab.moments.im/pkg"

// Forum ...
type Forum struct {
	ID int64 `db:"id"`
	// ForumID 发贴编号
	ForumID int64 `db:"forum_id"`
	// Type 帖子类型 同城 足球 电影。。。
	Type int32 `db:"type"`
	// ContentType 正文类别 1=表示发贴内容 2=表示回贴内容 3=表示回复评论 4=点赞 5=收藏 6=@提醒 用户回复是@另一用户，只能@自己好友
	ContentType int32 `db:"content_type"`
	// ViewCount 查看次数
	ViewCount int32 `db:"view_count"`
	// LikeCount 点赞次数
	LikeCount int64 `db:"like_count"`
	// CommentCount 回复次数
	CommentCount int64 `db:"comment_count"`
	// Content 发贴正文
	Content string `db:"content"`
	// Longitude 经度
	Longitude float32 `db:"longitude"`
	// Latitude 纬度
	Latitude float32 `db:"latitude"`
	// LocationCity 地理位置城市
	LocationCity string `db:"location_city"`
	// LocationName 地理位置名称
	LocationName string `db:"location_name"`
	// LocationAddress 地理位置名称详细名称
	LocationAddress string `db:"location_address"`
	// Status 查看权限1是公共2私有
	Status int32 `db:"status"`
	// CreateAt 发贴时间
	CreateAt int64 `db:"create_at"`
	// CreateBy 用户编号
	CreateBy int32 `db:"create_by"`
	// ContentEntity 内容解析@7 URL高亮
	ContentEntity string `db:"content_entity"`
	DeleteAt      int    `db:"delete_at"`
	Mention       string `db:"mention"` //提及
	Topic         string `db:"topic"`   //话题
	// 媒体文件
	Medias  []*Media
	Entitys []*pkg.HighLight
	// thumb
	IsThumb int32
	// 点赞用户
	ThumbUsers []int32
}

type ForumPageReq struct {
	CreateBy int32
	ForumID  int64
	TopicID  string
	Limit    int64
}

func (*Forum) TableName() string {
	return " forum "
}
