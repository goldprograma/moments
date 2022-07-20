package internal

import (
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// NoticeInsert 插入关系
func NoticeInsert(Notice *moment.Notice, db *sqlx.DB) error {
	return Insert(Notice, db, nil)
}

// NoticeGet 插入关系
func NoticeGet(Notice *moment.Notice, db *sqlx.DB) error {
	return Get(Notice, db)
}

// NoticeInsertBatch 插入关系
func NoticeInsertBatch(notices []*moment.Notice, db *sqlx.DB) (err error) {
	_, err = db.NamedExec(`insert into notice(
		notice_id,
		type,
		relation_id,
		notifier,
		status,
		create_at,
		create_by
		)VALUES(
			:notice_id,
			:type,
			:relation_id,
			:notifier,
			:status,
			:create_at,
			:create_by
			)`, notices)
	return err
}

// NoticeDelete 删除关系
func NoticeDelete(Notice *moment.Notice, db *sqlx.DB, tx *sqlx.Tx) (err error) {
	_, err = Delete(Notice, db, tx)
	return
}

// // 查询单个关系
// func GetOneNotice(Notice *models.Notice, db *sqlx.DB) error {
// 	return Get(Notice, db)
// }

// // NoticeUnReadCount 查询未读条数
// func NoticeUnReadCount(notice *models.Notice, db *sqlx.DB) (count int64, err error) {
// 	sql := "select count(1) from notice where notifier = ? and status = 0 "
// 	err = db.Get(&count, sql, notice.Notifier)
// 	return
// }

// NoticePage 查询消息列表
func NoticePage(notice *moment.NoticePageReq, db *sqlx.DB) (notices []*moment.Notice, err error) {
	sql := "select * from notice where status =1 and notifier = ? order by notice_id desc limit ?"
	if err = db.Select(&notices, sql, notice.Notifier, notice.Limit); err != nil {
		return nil, err
	}

	//清除未读

	if notice.HasRead {
		if _, err = db.Exec("update notice set status = 2 where notifier = ?", notice.Notifier); err != nil {
			return
		}
	}

	for _, notice := range notices {
		var content = &moment.NoticeContent{}
		switch notice.Type {
		case 1: // 点赞

			sql = "select 0 as comment_id, 0 as comment_content, a.forum_id, a.content as forum_content,b.name as media_content from  forum a  left join media b on b.main_id = a.forum_id where  a.forum_id = ? and b.seq = 0"
			if err = db.Get(content, sql, notice.RelationID); err != nil {
				return
			}

		case 2: //评论
			sql = "select a.comment_id,a.content as comment_content,  b.forum_id, b.content as forum_content,c.name as media_content  from comment a  left join forum b on a.forum_id = b.forum_id left join media c on c.main_id = b.forum_id  where a.comment_id = ? and c.seq = 0"
			if err = db.Get(content, sql, notice.RelationID); err != nil {
				return
			}

		}
		notice.Content = content

	}
	return
}
