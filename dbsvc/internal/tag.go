package internal

import (
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

func TagInsert(tag *moment.Tag, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(tag, db, tx)
}
func UserTagInsert(usertag []*moment.UserTag, db *sqlx.DB, tx *sqlx.Tx) (err error) {
	_, err = tx.NamedExec(`insert into user_tag(
user_tag_id,
tag_id,
user_id,
create_at,
create_by
	) VALUES (
		:user_tag_id,
		:tag_id,
		:user_id,
		:create_at,
		:create_by
		)`, usertag)
	return
}

func TagGet(userID int32, db *sqlx.DB) (tags []*moment.Tag, err error) {
	sql := "select * from tag where create_by= ? "
	err = db.Select(&tags, sql, userID)
	return
}

func UserTagGet(userID int32, tagID int64, db *sqlx.DB) (tags []*moment.UserTag, err error) {
	sql := "select * from user_tag where user_id= ? and tag_id = ?"
	err = db.Select(&tags, sql, userID, tagID)
	return
}
