package datamodel

// ForumIgnore 忽略动态(仅忽略系统推荐动态，用户动态可以直接删除)
type ForumIgnore struct {
	ID       int `db:"id" json:"-"`
	IgnoreID int `db:"ignore_id" json:"ignore_id"`
	ForumID  int `db:"forum_id" json:"forum_id"`
	CreateBy int `db:"create_by" json:"create_by"`
	CreateAt int `db:"create_at" json:"create_at"`
}
