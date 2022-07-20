package internal

import (
	"fmt"
	"strconv"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

//CommentInsert 插入评论
func CommentInsert(comment *moment.Comment, db *sqlx.DB, tx *sqlx.Tx) error {
	return Insert(comment, db, tx)
}

// CommentThumbUpdate 更新帖子点赞数
func CommentThumbUpdate(commentID, thumbup int64, db *sqlx.DB) (err error) {
	sqlStr := "update comment set thumb_up=thumb_up+? where comment_id = ?"
	_, err = db.Exec(sqlStr, thumbup, commentID)
	return
}

//CommentSubCommentCountUpdate 更新帖子ID
func CommentSubCommentCountUpdate(commentID int64, commentCount int64, db *sqlx.DB) error {
	_, err := db.Exec("update comment set sub_comments = sub_comments+? where comment_id = ?", commentCount, commentID)
	return err
}

// CommentDelete 删除评论
func CommentDelete(comment *moment.Comment, db *sqlx.DB) (err error) {
	sqlStr := "delete from  comment where  comment_id= ? or sup_id= ?"
	_, err = db.Exec(sqlStr, comment.CommentID, comment.CommentID)
	return err
}

//CommentGet 查询单个评论
func CommentGet(comment *moment.Comment, db *sqlx.DB) error {
	return Get(comment, db)
}

//CommentAllPage 分页查询
func CommentAllPage(req *moment.CommentPageReq, db *sqlx.DB) ([]*moment.Comment, error) {
	sql := "select * from comment  where forum_id = ? and comment_id > ?"
	var limitSQL = ""
	if req.Limit != 0 {
		limitSQL = " limit " + strconv.Itoa(int(req.Limit))
	}
	var sortSQL = " Order by comment_id asc "
	if req.Order != "" {
		if req.Sort == "" {
			req.Sort = " desc "
		}
		sortSQL = " Order by" + req.Order + req.Sort
	}

	comments := make([]*moment.Comment, 0, req.Limit)
	sql += sortSQL + limitSQL
	err := db.Select(&comments, sql, req.ForumID, req.CommentID)
	if err != nil {
		return nil, err
	}
	return comments, err
}

//CommentPageOrderByThumbup 分页查询
func CommentPageOrderByThumbup(req *moment.CommentPageReq, db *sqlx.DB) ([]*moment.Comment, error) {

	sql := "select * from comment a where a.forum_id = ? and a.sup_id is null "

	var commentSQL = ""
	if req.CommentID > 0 {
		commentSQL = fmt.Sprintf(" and comment_id  < %d", req.CommentID)
	}
	sql += commentSQL + " order by a.thumb_up desc,a.create_at asc"

	var limitSQL = ""
	if req.Limit != 0 {
		limitSQL = " limit " + strconv.Itoa(int(req.Limit))
	}

	comments := make([]*moment.Comment, 0, req.Limit)
	sql += limitSQL
	err := db.Select(&comments, sql, req.ForumID)
	if err != nil {
		return nil, err
	}

	return comments, err
}

//CommentPage 分页查询
func CommentPage(req *moment.CommentPageReq, db *sqlx.DB) ([]*moment.Comment, error) {
	sql := "select * from comment a  where forum_id = ? and sup_id is null "
	var limitSQL = ""
	if req.Limit != 0 {
		limitSQL = " limit " + strconv.Itoa(int(req.Limit))
	}
	var sortSQL = " Order by comment_id desc "

	if req.Sort == "" || req.Sort == "desc" {
		if req.CommentID > 0 {
			sql += fmt.Sprintf(" and comment_id < %d", req.CommentID)
		}
	} else {
		sql += fmt.Sprintf(" and comment_id > %d", req.CommentID)
		sortSQL = " Order by" + req.Order + req.Sort
	}

	comments := make([]*moment.Comment, 0, req.Limit)
	sql += sortSQL + limitSQL
	err := db.Select(&comments, sql, req.ForumID)
	if err != nil {
		return nil, err
	}
	return comments, err
}

//CommentReplayCount 好友评论数
func CommentReplayCount(req *moment.ReplayPageReq, db *sqlx.DB) (count int64, err error) {

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
	sql := "select count(1) from comment  where sup_id = ?" + friendSQL

	if err = db.Get(&count, sql, req.SupID); err != nil {
		return 0, err
	}

	return
}

//ReplayCommentPage 分页查询
func ReplayCommentPage(req *moment.ReplayPageReq, db *sqlx.DB) ([]*moment.Comment, error) {
	sql := "select * from comment  where forum_id = ? and comment_id > ?  and sup_id = ? "
	userMap, err := UserIgnores(req.CreateBy, db)
	if err != nil {
		return nil, err
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
	var limitSQL = ""
	if req.Limit != 0 {
		limitSQL = " limit " + strconv.Itoa(int(req.Limit))
	}
	var sortSQL = " Order by comment_id asc "
	if req.Order != "" {
		if req.Sort == "" {
			req.Sort = " desc "
		}
		sortSQL = " Order by" + req.Order + req.Sort
	}

	comments := make([]*moment.Comment, 0, req.Limit)
	sql += friendSQL + sortSQL + limitSQL
	err = db.Select(&comments, sql, req.ForumID, req.CommentID, req.SupID)
	if err != nil {
		return nil, err
	}
	return comments, err
}
