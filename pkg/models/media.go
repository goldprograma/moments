package models

// Media ...
type Media struct {
	ID int64 `db:"id"`
	// ForumID 发贴编号
	MainID int64 `db:"main_id"`
	// Seq 图片序号
	Seq int64 `db:"seq"`
	// Name 文件名
	Name string `db:"name"`
	// Ext 附件类型 1=表示图片(bmp、jpg、jpeg、png) 2=视频(gif、mp4等) 3=音频 4=DOC文件pdf,txt, doc,docx, xls, ppt 等 5=压缩文件(zip,rar,tar) 6=需要获取其它数据URL 7=其它
	Ext int32 `db:"ext"`
	// Thum 缩略图信息
	Thum string `db:"thum"`
	// Region s3 区域 以后可能需要分区域存储
	Region string `db:"region"`
	// Size 文件大小
	Size int32 `db:"size"`
	// ThumSize 缩略图大小
	ThumSize int32 `db:"thum_size"`
	// Hash 换算出来的文件的hash 值
	Hash string `db:"hash"`
	// Duration 视频文件时长
	Duration int32 `db:"duration"`
	// Height 图片尺寸
	Height int32 `db:"height"`
	// Width 图片尺寸
	Width int32 `db:"width"`
	// CreateTime 创建时间
	CreateAt int64 `db:"create_at"`
}

func (*Media) TableName() string {
	return " media "
}
