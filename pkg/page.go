package pkg

import "gitlab.moments.im/pkg/protoc/moment"

type PageRows struct {
	Page *moment.Page
	Rows interface{}
}

type RowsModel interface {
	Insert()
	Update()
	Get()
	Delete()
}

type Page struct {
	// 总条数
	TotalRows int64 `url:"totalrows"`
	// 总页数
	PageCount int64 `url:"pagecount"`
	// 页数大小
	PageSize int64 `url:"pagesize"`
	// 当前页
	CurrentPage int64 `url:"currentpage"`
	// 偏移量
	Offset int64 `url:"offset"`
	// 长度
	Limit int64 `url:"limit"`
	// 排序字段
	Order string `url:"order"`
	// 排序方法 (AES,DESC)
	Sort string `url:"sort"`
}
