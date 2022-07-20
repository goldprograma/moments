package internal

import (
	"fmt"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// ForumsTopicPage 查询话题帖子
func ForumsTopicPage(req *moment.ForumTopicPageReq, db *sqlx.DB) (forums []*moment.ForumTopic, err error) {
	// var friendSQL = ""
	// for _, friend := range req.Friends {
	// 	friendSQL += "," + strconv.Itoa(int(friend))
	// }
	// if friendSQL != "" {
	// 	friendSQL = " and a.create_by in(" + friendSQL[1:] + ") "
	// }
	var offsetSQL = ""
	if req.ForumID > 0 {
		offsetSQL = fmt.Sprintf(" and a.forum_id < %d ", req.ForumID)
	}
	sql := `
	select * from forum_topic a  where topic_id = ? ` + offsetSQL + `
	AND NOT EXISTS (select forum_id from forum_ignore e where a.forum_id = e.forum_id and  e.create_by =? )
ORDER BY
	a.forum_id DESC 
	LIMIT ?
	`

	err = db.Select(&forums, sql, req.TopicID, req.UserID, req.Limit)
	return
}

// ForumTopicGet 查询单个关系
func ForumTopicGet(forum *moment.ForumTopic, db *sqlx.DB) error {
	return Get(forum, db)
}

// ForumTopicInsert 查询单个关系
func ForumTopicInsert(forum *moment.ForumTopic, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(forum, db, tx)
}
