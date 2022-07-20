package internal

import (
	"database/sql"
	"fmt"
	"strconv"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

type ForumUserThumbCount struct {
	UserID int32 `db:"user_id"`
	Count  int64 `db:"count"`
}

// ThumbInsert 插入关系
func ThumbInsert(Thumb *moment.Thumb, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(Thumb, db, tx)
}

// ThumbUpdate 修改关系
func ThumbUpdate(Thumb *moment.Thumb, db *sqlx.DB, tx *sqlx.Tx, whereCols ...string) error {
	return Update(Thumb, db, tx, whereCols...)
}

// ThumbCommentDelete 删除评论
func ThumbCommentDelete(thumb *moment.Thumb, db *sqlx.DB) (count int64, err error) {
	sqlStr := "delete from thumb where  comment_id= ? "
	var sqlResult sql.Result

	if sqlResult, err = db.Exec(sqlStr, thumb.CommentID); err != nil {
		return 0, err
	}

	return sqlResult.RowsAffected()
}

// ThumbForumDelete 删除评论
func ThumbForumDelete(thumb *moment.Thumb, db *sqlx.DB) (count int64, err error) {
	sqlStr := "delete from thumb where  forum_id= ?"
	var sqlResult sql.Result

	if sqlResult, err = db.Exec(sqlStr, thumb.ForumID); err != nil {
		return 0, err
	}

	return sqlResult.RowsAffected()
}

// ThumbCommentUserDelete 删除评论
func ThumbCommentUserDelete(thumb *moment.Thumb, db *sqlx.DB) (count int64, err error) {
	sqlStr := "delete from thumb where  comment_id= ? and create_by = ?"
	var sqlResult sql.Result

	if sqlResult, err = db.Exec(sqlStr, thumb.CommentID, thumb.CreateBy); err != nil {
		return 0, err
	}

	return sqlResult.RowsAffected()
}

// ThumbForumUserDelete 删除评论
func ThumbForumUserDelete(thumb *moment.Thumb, db *sqlx.DB) (count int64, err error) {
	sqlStr := "delete from thumb where  forum_id= ? and comment_id=0 and create_by = ?"
	var sqlResult sql.Result

	if sqlResult, err = db.Exec(sqlStr, thumb.ForumID, thumb.CreateBy); err != nil {
		return 0, err
	}

	return sqlResult.RowsAffected()
}

// ThumbForumGet 查询单个点赞
// func ThumbForumGet(req *moment.Thumb, db *sqlx.DB) error {
// 	return db.Get(req, "select * from thumb where forum_id = ? ", req.ForumID)
// }

// ThumbCommentGet 查询回复单个点赞
// func ThumbCommentGet(req *moment.Thumb, db *sqlx.DB) error {
// 	return db.Get(req, "select * from thumb where comment_id = ? ", req.CommentID)
// }

//ThumbUserCount 统计
func ThumbUserCount(req *moment.ThumbUserCountReq, db *sqlx.DB) (count int64, err error) {
	err = db.Get(&count, "select count(1) from thumb where forum_uid = ? or comment_uid=? ", req.UserID, req.UserID)
	return
}

// ThumbPage 根据条件查询
func ThumbPage(thumb *moment.ThumbPageReq, db *sqlx.DB) (thumbs []*moment.Thumb, err error) {
	var upDown string
	if thumb.UpDown > 0 {
		upDown = fmt.Sprintf(" and up_down = %d", thumb.UpDown)
	}
	var limitSQL = ""
	if thumb.Limit > 0 {
		limitSQL = fmt.Sprintf(" limit %d", thumb.Limit)
	}
	var offSet string
	if thumb.ThumbID > 0 {
		offSet = fmt.Sprintf(" and thumb_id < %d ", thumb.ThumbID)
	}
	sql := "select * from thumb where forum_id = ? " + offSet + " and comment_id = 0  " + upDown + " order by thumb_id desc " + limitSQL

	err = db.Select(&thumbs, sql, thumb.ForumID)
	return thumbs, err
}

// ThumbAll 查询所有数据
func ThumbAll(Thumb *moment.Thumb, db *sqlx.DB) ([]*moment.Thumb, error) {
	sql := "select * from thumb where forum_id=?"
	Thumbs := make([]*moment.Thumb, 0)
	err := db.Select(&Thumbs, sql, Thumb.ForumID)
	return Thumbs, err
}

// ThumbUsers 查询所有点赞用户
func ThumbUsers(createBy int32, forumID int64, friends []int32, db *sqlx.DB) ([]int32, error) {
	userMap, err := UserIgnores(createBy, db)
	if err != nil {
		return nil, err
	}
	var friendSQL = ""
	for _, friend := range friends {
		if _, ok := userMap[friend]; !ok {
			friendSQL += "," + strconv.Itoa(int(friend))
		}
	}
	if friendSQL != "" {
		friendSQL = fmt.Sprintf(" AND create_by in(%s)", friendSQL[1:])
	}
	sql := "select create_by from thumb where forum_id=? " + friendSQL
	users := make([]int32, 0)
	err = db.Select(&users, sql, forumID)
	return users, err
}

// ThumbForumUsers 查询所有点赞用户
func ThumbForumUsers(forumID, thumbID int64, limit int32, db *sqlx.DB) ([]int32, error) {
	var limitSQL = ""
	if limit > 0 {
		limitSQL = fmt.Sprintf(" limit %d", limit)
	}
	var offSet string
	if thumbID > 0 {
		offSet = fmt.Sprintf(" and thumb_id < %d ", thumbID)
	}
	sql := "select create_by from thumb where forum_id=? " + offSet + " order by thumb_id asc  " + limitSQL
	users := make([]int32, 0)
	err := db.Select(&users, sql, forumID)
	return users, err
}

// ThumbForumCheck 检查用户是否点赞
func ThumbForumCheck(forumID int64, createBy int32, db *sqlx.DB) (has bool, err error) {
	sql := "select if(count(1)>0,true,false)  from thumb where forum_id=? and comment_id = 0 and create_by = ?"
	err = db.Get(&has, sql, forumID, createBy)
	return has, err
}

// ThumbCommentCheck 检查用户是否点赞
func ThumbCommentCheck(commentID int64, createBy int32, db *sqlx.DB) (has bool, err error) {
	sql := "select if(count(1)>0,true,false) from thumb where comment_id=? and create_by = ?"
	err = db.Get(&has, sql, commentID, createBy)
	return has, err
}

//ThumbCommentCount 获取评论点赞数量
func ThumbCommentCount(req *moment.ThumbPageReq, db *sqlx.DB) (count int64, err error) {
	userMap, err := UserIgnores(req.CreateBy, db)
	if err != nil {
		return 0, err
	}
	var friendSQL = ""
	for _, friend := range req.Friends {
		if _, ok := userMap[friend]; !ok {
			friendSQL += "," + strconv.Itoa(int(friend))
		}
	}
	if friendSQL != "" {
		friendSQL = fmt.Sprintf(" AND create_by in(%s)", friendSQL[1:])
	}

	sql := "select count(1) from thumb  where  comment_id = ? " + friendSQL
	if err = db.Get(&count, sql, req.CommentID); err != nil {
		return 0, err
	}
	return
}

//ThumbForumCount 获取帖子的点赞数量
func ThumbForumCount(req *moment.ThumbPageReq, db *sqlx.DB) (count int64, err error) {
	userMap, err := UserIgnores(req.CreateBy, db)
	if err != nil {
		return 0, err
	}
	var friendSQL = ""
	for _, friend := range req.Friends {
		if _, ok := userMap[friend]; !ok {
			friendSQL += "," + strconv.Itoa(int(friend))
		}
	}
	if friendSQL != "" {
		friendSQL = fmt.Sprintf(" AND create_by in(%s)", friendSQL[1:])
	}

	sql := "select count(1) from thumb  where forum_id = ? and comment_id = 0 " + friendSQL
	if err = db.Get(&count, sql, req.ForumID); err != nil {
		return 0, err
	}
	return
}

// ThumbForumUserGet  查询帖子
func ThumbForumUserGet(forumID int64, db *sqlx.DB) (usercount []*ForumUserThumbCount, err error) {
	sqlStr := `select forum_uid as user_id,count(forum_uid) as count from thumb  where forum_id = ? and comment_id = 0 GROUP BY forum_uid
	UNION
	select comment_uid as user_id,count(comment_uid) as count from thumb  where forum_id = ? and comment_id > 0 GROUP BY comment_uid
	`
	err = db.Select(&usercount, sqlStr, forumID, forumID)
	return
}

// ThumbCommentUserGet  查询评论帖子
func ThumbCommentUserGet(commentID int64, db *sqlx.DB) (usercount []*ForumUserThumbCount, err error) {
	sqlStr := `select count(1) as count from thumb  where comment_id =?`
	err = db.Select(&usercount, sqlStr, commentID)
	return
}
