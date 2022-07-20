package models

// ForumType ...
type ForumType struct {
	// ID 主键
	ID int32 `db:"id"`
	// Name 类型
	Name string `db:"name"`
	// CreateAt 创建时间
	CreateAt int64 `db:"create_at"`
	// UpdateAt 修改时间
	UpdateAt int64 `db:"update_at"`
}

func (*ForumType) TableName() string {
	return " forum_type "
}
