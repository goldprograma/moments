package internal

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// FollowInsert 插入关系
func FollowInsert(follow *moment.Follow, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(follow, db, tx)
}

// FollowUpdate 修改关系
func FollowUpdate(Follow *moment.Follow, db *sqlx.DB, tx *sqlx.Tx, whereCols ...string) error {
	return Update(Follow, db, tx, whereCols...)
}

// FollowDelete 删除关系
func FollowDelete(Follow *moment.Follow, db *sqlx.DB, tx *sqlx.Tx) error {
	result, err := Delete(Follow, db, tx)
	if err != nil {
		return err
	}
	if count, _ := result.RowsAffected(); count == 0 {
		return errors.New("没有关注信息")
	}
	return nil
}

// FollowGet 查询单个关系
func FollowGet(Follow *moment.Follow, db *sqlx.DB) error {
	return Get(Follow, db)
}

// FollowCount 我关注的
func FollowCount(page *moment.FollowCountReq, db *sqlx.DB) (count int64, err error) {

	sql := "select count(1) from follow where create_by= ? "
	err = db.Get(&count, sql, page.UserID)
	return
}

// FollowFansCount 关注我的
func FollowFansCount(page *moment.FollowCountReq, db *sqlx.DB) (count int64, err error) {
	sql := "select count(1) from follow where follow_uid= ?  "
	err = db.Get(&count, sql, page.UserID)
	return
}

func FollowByMe(page *moment.FollowPageReq, db *sqlx.DB) (follows []*moment.Follow, err error) {

	sql := "select * from follow where create_by= ?  order by follow_id desc limit ?,?"
	err = db.Select(&follows, sql, page.CreateBy, page.Offset, page.Limit)
	return
}

//FollowCheck 是否已关注
func FollowCheck(createBy, followUID int32, db *sqlx.DB) (has bool, err error) {
	sql := "select if(count(1)>0,true,false) from follow where create_by= ?  and follow_uid = ?"
	err = db.Get(&has, sql, createBy, followUID)
	return
}

// FollowMe 关注我的
func FollowMe(page *moment.FollowPageReq, db *sqlx.DB) (follows []*moment.Follow, err error) {
	sql := "select * from follow where follow_uid= ?  order by follow_id desc limit ?,?"
	err = db.Select(&follows, sql, page.CreateBy, page.Offset, page.Limit)
	return
}

//FollowMeAll
func FollowMeAll(followUid int32, db *sqlx.DB) (follows []*moment.Follow, err error) {
	sql := "select * from follow where follow_uid= ?"
	err = db.Select(&follows, sql, followUid)
	return
}

//FansAll 我的所有粉丝
func FansAll(followUid int32, db *sqlx.DB) (follows []*moment.Follow, err error) {
	sql := "select * from follow where follow_uid= ?"
	err = db.Select(&follows, sql, followUid)
	return
}

//FollowAll 我的所有关注
func FollowAll(createBy int32, db *sqlx.DB) (follows []*moment.Follow, err error) {
	sql := "select * from follow where create_by= ? order by follow_id desc"
	err = db.Select(&follows, sql, createBy)
	return
}

//FollowAllOrderByCreateAt 我的所有关注过滤用户
func FollowAllOrderByCreateAt(createBy int32, crateAt int32, db *sqlx.DB) (follows []*moment.Follow, err error) {
	var orderSQL string
	if crateAt > 0 {
		orderSQL = fmt.Sprintf("and create_at< %d", crateAt)
	}
	sql := "select * from follow where create_by= ?  " + orderSQL + " order by create_at desc"
	err = db.Select(&follows, sql, createBy)
	return
}

//FollowMeAllID 我的所有粉丝ID
func FollowMeAllID(followUID int32, db *sqlx.DB) (ids []int32, err error) {
	sql := "select create_by from follow where follow_uid= ?"
	err = db.Select(&ids, sql, followUID)
	return
}

//FollowFansIDPage 我的所有粉丝ID
func FollowFansIDPage(req *moment.FansIDReq, db *sqlx.DB) (ids []int32, err error) {
	sql := "select create_by from follow where follow_uid= ? order by follow_id desc limit ?,?"
	err = db.Select(&ids, sql, req.UserID, req.Offset, req.Limit)
	return
}

//FollowAllID 我的所有粉丝ID
func FollowAllID(createBy int32, db *sqlx.DB) (ids []int32, err error) {
	sql := "select  follow_uid from follow where create_by= ? order by follow_id desc"
	err = db.Select(&ids, sql, createBy)
	return
}

//FollowAllIDAndTime 我的所有粉丝ID
// func FollowAllIDAndTime(req *moment.FollowFourmPageReq, db *sqlx.DB) (follows map[int32]int64, err error) {
// 	follows = make(map[int32]int64, 0)
// 	sql := "select  follow_uid,create_at from follow where create_by= ? order by follow_id desc"
// 	// if req.FollowUID > 0 {
// 	// 	sql += fmt.Sprintf(" and follow_uid < %d", req.FollowUID)
// 	// }
// 	// sql +=
// 	err = db.Select(&follows, sql, req.UserID)
// 	return
// }

// FollowInsertBatch 新增批量
func FollowInsertBatch(follows []*moment.Follow, db *sqlx.DB) error {
	_, err := db.NamedExec("insert into follow (follow_id,create_by,follow_uid,create_at) values (:follow_id,:create_by,:follow_uid,:create_at)", follows)
	return err
}

// FollowDeleteBatch 删除批量
func FollowDeleteBatch(follows []*moment.Follow, db *sqlx.DB) error {
	sql := "delete from follow where create_by= ? and  follow_uid in ("
	for _, v := range follows {
		id := strconv.Itoa(int(v.FollowUID))
		sql = sql + id + ", "
	}
	sqlnew := strings.TrimSuffix(sql, ", ")
	sqlnew = sqlnew + ")"
	_, err := db.Exec(sqlnew, follows[0].CreateBy)
	return err
}

//FansCountBySource 我的所有关注
func FansCountBySource(createBy int32, source int64, db *sqlx.DB) (count int32, err error) {
	sql := "select count(*) from follow where follow_uid= ? and follow_source=?"
	err = db.Get(&count, sql, createBy, source)
	return
}
