package pkg

//NoticeType 通知消息类型
type NoticeType int

const (
	//NoticeTypeLike 1点赞
	NoticeTypeLike NoticeType = 1 << iota
	//NoticeTypeComment 2评论
	NoticeTypeComment
	//NoticeTypeFollow  3 关注
	NoticeTypeFollow
	//NoticeTypeAt 4 @
	NoticeTypeAt
	//NoticeTypeMention  5 提及
	NoticeTypeMention
)
