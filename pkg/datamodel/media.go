package datamodel

// Media model
type Media struct {
	ID       int64  `db:"id" json:"-"`
	MainID   int64  `db:"main_id" json:"MainID"`      //动态forum_id
	Seq      int64  `db:"seq" json:"Seq"`             //序号
	Name     string `db:"name" json:"Name,omitempty"` //文件名
	Ext      int    `db:"ext" json:"Ext"`             //类型
	Thum     string `db:"thum" json:"Thum"`           //缩略图
	Region   string `db:"region" json:"Region,omitempty"`
	Size     int    `db:"size" json:"Size,omitempty"`
	ThumSize int    `db:"thum_size" json:"ThumSize,omitempty"`
	Hash     string `db:"hash" json:"Hash,omitempty"`
	Duration int    `db:"duration" json:"Duration,omitempty"`
	Height   int    `db:"height" json:"Height,omitempty"`
	Width    int    `db:"width" json:"Width,omitempty"`
	CreateAt int    `db:"create_at" json:"CreateAt,omitempty"`
}
