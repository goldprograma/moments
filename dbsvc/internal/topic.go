package internal

import (
	"database/sql"
	"time"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

//TopicInsert 插入评论
func TopicInsert(Topic *moment.Topic, db *sqlx.DB) error {
	return Insert(Topic, db, nil)
}

// TopicDelete 删除评论
func TopicDelete(Topic *moment.Topic, db *sqlx.DB) (count int64, err error) {
	sqlStr := "update  topic set status = 2 and topic_id= ?"
	var sqlResult sql.Result
	if sqlResult, err = db.Exec(sqlStr, time.Now().Unix, Topic.TopicID); err != nil {
		return 0, err
	}
	return sqlResult.RowsAffected()
}

//TopicPage 获取所有的话题
func TopicPage(req *moment.TopicPageReq, db *sqlx.DB) (topcs []*moment.Topic, err error) {
	topcs = make([]*moment.Topic, 0, req.Limit)
	var args = make([]interface{}, 0, 3)
	var where string
	args = append(args, req.TopicID)
	if req.TopicName != "" {
		where += " and topic_name like ?"
		args = append(args, "%"+req.TopicName+"%")
	}
	args = append(args, req.Limit)
	sql := "select * from topic  where status = 1 and topic_id > ?  " + where + "  limit ?"
	err = db.Select(&topcs, sql, args...)
	return
}

//TopicTypeAll 获取所有的话题类型
func TopicTypeAll(db *sqlx.DB) (topcTypes []*moment.TopicType, err error) {
	topcTypes = make([]*moment.TopicType, 0)
	sql := "select * from topic_type where status = 1 order by seq asc "
	err = db.Select(&topcTypes, sql)
	return
}
