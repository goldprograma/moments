package internal

import (
	"fmt"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// ForumRecommendDelete 删除关系

// ForumRecommendGet 查询单个关系
func ForumRecommendGet(forum *moment.ForumRecommend, db *sqlx.DB) error {
	return Get(forum, db)
}

//ForumRecommendInsert 插入关系
func ForumRecommendInsert(forum *moment.ForumRecommend, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(forum, db, tx)
}

// ForumRecommendPage 查询推荐帖子
func ForumRecommendPage(req *moment.RecommendPageReq, db *sqlx.DB) (forums []*moment.ForumRecommend, err error) {
	var offsetSQL = ""
	if req.ForumID > 0 {
		offsetSQL = fmt.Sprintf(" and a.forum_id < %d", req.ForumID)
	}
	sql := `
	select * from forum_recommend a  where
	NOT EXISTS (select forum_id from forum_ignore e where a.forum_id = e.forum_id and  e.create_by =? )   ` + offsetSQL + `
ORDER BY
	a.forum_id DESC 
	LIMIT ?
	`
	err = db.Select(&forums, sql, req.UserID, req.Limit)
	return
}
