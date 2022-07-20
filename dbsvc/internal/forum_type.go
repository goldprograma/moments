package internal

import (
	"gitlab.moments.im/pkg/models"

	"github.com/jmoiron/sqlx"
)

// 插入关系
func InsertForumType(ForumType *models.ForumType, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(ForumType, db, tx)
}

// 修改关系
func UpdateForumType(ForumType *models.ForumType, db *sqlx.DB, tx *sqlx.Tx, whereCols ...string) error {
	return Update(ForumType, db, tx, whereCols...)
}

// DeleteForumType 删除关系
func DeleteForumType(ForumType *models.ForumType, db *sqlx.DB, tx *sqlx.Tx) (err error) {
	_, err = Delete(ForumType, db, tx)
	return
}

// 查询单个关系
func GetOneForumType(ForumType *models.ForumType, db *sqlx.DB) error {
	return Get(ForumType, db)
}

// 查询所有数据
func GetAllForumTypes(ForumType *models.ForumType, db *sqlx.DB) ([]*models.ForumType, error) {
	sql := "select * from " + ForumType.TableName()
	ForumTypes := make([]*models.ForumType, 0)
	err := db.Select(&ForumTypes, sql)
	return ForumTypes, err
}
