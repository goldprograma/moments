package internal

import (
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

//UserIgnoreGet 获取忽略配置
func UserIgnoreGet(req *moment.UserIgnore, db *sqlx.DB) ([]*moment.UserIgnore, error) {
	rep := make([]*moment.UserIgnore, 0)
	var err error
	sql, args := GetConditionSQL(req, db)
	err = db.Select(&rep, "select * from user_ignore "+sql, args...)
	return rep, err
}

//UserIgnoreAdd 新增忽略
func UserIgnoreAdd(ignores []*moment.UserIgnore, db *sqlx.DB) (err error) {
	return InsertBatch(ignores, db, nil)
}

//UserIgnoreDelete 更新
func UserIgnoreDelete(ignore *moment.UserIgnore, db *sqlx.DB) (err error) {

	_, err = db.Exec("delete from user_ignore where   look = ? AND  user_id = ? AND  create_by = ?", ignore.Look, ignore.UserID, ignore.CreateBy)
	return err
}

//UserIgnoreCheck 忽略配置
func UserIgnoreCheck(req *moment.UserIgnore, db *sqlx.DB) (bool, error) {
	var err error
	var count int32
	sql := "select count(1) from user_ignore a where  a.user_id =? and a.create_by =? and a.look = 1"
	err = db.Get(&count, sql, req.CreateBy, req.UserID)
	return count > 0, err
}

//UserIgnores 忽略配置
func UserIgnores(createBy int32, db *sqlx.DB) (map[int32]struct{}, error) {
	var err error
	var users []int32
	sql := `select user_id  from user_ignore where  create_by =? and look = 2 
	UNION
	select create_by  from user_ignore where  user_id =? and look = 1`
	if err = db.Select(&users, sql, createBy, createBy); err != nil {
		return nil, err
	}
	var userMap = make(map[int32]struct{})
	for _, id := range users {
		userMap[id] = struct{}{}
	}
	return userMap, err
}

//UserIgnoresAllID 忽略配置
func UserIgnoresAllID(createBy int32, db *sqlx.DB) ([]int32, error) {
	var err error
	var users []int32
	sql := `select user_id from user_ignore where  create_by =? and look = 2 
	UNION
	select create_by  from user_ignore where  user_id =? and look = 1`
	if err = db.Select(&users, sql, createBy, createBy); err != nil {
		return nil, err
	}
	return users, err
}

//UserIgnoreMeGet 获取忽略我得
func UserIgnoreMeGet(ig *moment.UserIgnore, db *sqlx.DB) ([]int32, error) {
	var err error
	var users []int32
	sql := `select create_by  from user_ignore where   user_id=? and look = 1`
	if err = db.Select(&users, sql, ig.UserID); err != nil {
		return nil, err
	}
	return users, err
}

// IgnoreForum 不看该帖子
func IgnoreForum(forum *moment.ForumIgnore, db *sqlx.DB) (err error) {
	err = Insert(forum, db, nil)
	return
}
