package internal

import (
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// RecommendInsert 插入关系
func RecommendInsert(Recommend *moment.UserRecommend, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(Recommend, db, tx)
}

// RecommendUpdate 修改关系
func RecommendUpdate(Recommend *moment.UserRecommend, db *sqlx.DB, tx *sqlx.Tx, whereCols ...string) error {
	return Update(Recommend, db, tx, whereCols...)
}

// RecommendDelete 删除关系
func RecommendDelete(Recommend *moment.UserRecommend, db *sqlx.DB, tx *sqlx.Tx) (err error) {
	_, err = Delete(Recommend, db, tx)
	return
}

// RecommendGet 查询单个关系
func RecommendGet(Recommend *moment.UserRecommend, db *sqlx.DB) error {
	return Get(Recommend, db)
}

// // RecommendPage 根据条件查询
// func RecommendPage(Recommend *moment.UserRecommend, page *moment.Page, db *sqlx.DB) (*moment.UserRecommendPageRep, error) {
// 	Forums := make([]*moment.UserRecommend, 0)
// 	cond, args := GetConditionSQL(Recommend, db)
// 	p := &pkg.PageRows{
// 		Page: page,
// 		Rows: &Forums,
// 	}
// 	sql := "select * from user_recommend " + cond
// 	err := GetPageRows(p, sql, db, args...)
// 	return &moment.UserRecommendPageRep{Page: page, Rows: Forums}, err
// }

// RecommendAll 查询所有数据
func RecommendAll(db *sqlx.DB) ([]*moment.UserRecommend, error) {
	sql := "select * from user_recommend "
	Recommends := make([]*moment.UserRecommend, 0)
	err := db.Select(&Recommends, sql)
	return Recommends, err
}

// RecommendAllUser 查询所有数据
func RecommendAllUser(db *sqlx.DB) ([]string, error) {
	sql := "select user_id from user_recommend "
	users := make([]string, 0)
	err := db.Select(&users, sql)
	return users, err
}
