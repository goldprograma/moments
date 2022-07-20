package models

// Topic ...
type Topic struct {
	// ID 主键
	ID int32 `db:"id"`
	// TopicID ID
	TopicID int64 `db:"topic_id"`
	// TopicName 话题
	TopicName string `db:"topic_name" select:"like"`
	// TypeID 类型ID
	TypeID int8 `db:"type_id"`
	// CreateAt 类型名称
	CreateAt int32 `db:"create_at"`
	// CreateBy 创建人
	CreateBy int32 `db:"create_by"`
	// Status 状态
	Status int8 `db:"status"`
}

// TopicType ...
type TopicType struct {
	// ID 主键
	ID int32 `db:"id"`
	// TopicTypeID ID
	TopicTypeID int64 `db:"topic_type_id"`
	// TopicTypeName 名称
	TopicTypeName string `db:"topic_type_name"`
	// CreateAt 创建事件
	CreateAt int32 `db:"create_at"`
	// CreateBy 创建人
	CreateBy int32 `db:"create_by"`
	// Seq 排序
	Seq int8 `db:"seq"`
	// Status 状态
	Status int8 `db:"status"`
}

type TopicPageReq struct {
	TopicTypeID int64
	TopicName   string
	TopicID     int64
	Limit       int64
	CreateBy    int32
}
