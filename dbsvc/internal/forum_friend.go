package internal

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/jmoiron/sqlx"
)

// ForumFriendPrivateSet 设置帖子私密帖子
func ForumPrivateSet(forum *moment.ForumFriend, db *sqlx.DB) (count int64, err error) {

	sqlStr := "update forum_friend set permission = ? where forum_id = ? and create_by = ?"
	var result sql.Result
	if result, err = db.Exec(sqlStr, forum.Permission, forum.ForumID, forum.CreateBy); err != nil {
		return
	}
	count, err = result.RowsAffected()
	return
}

// ForumDelete 删除帖子
func ForumDelete(forum *moment.ForumFriend, db *sqlx.DB) (count int64, err error) {
	tx, err := db.Beginx()
	var result sql.Result
	if err != nil {
		return 0, err
	}
	if result, err = tx.Exec("delete from forum_friend where forum_id=? ", forum.ForumID); err != nil {
		tx.Rollback()
		return 0, err
	}
	if _, err = tx.Exec("delete from forum_recommend where forum_id=? ", forum.ForumID); err != nil {
		tx.Rollback()
		return 0, err
	}
	if _, err = tx.Exec("delete from forum_topic where forum_id=? ", forum.ForumID); err != nil {
		tx.Rollback()
		return 0, err
	}

	tx.Commit()

	return result.RowsAffected()
}

// ForumUserDelete 删除帖子
func ForumUserDelete(forum *moment.ForumFriend, db *sqlx.DB) (err error) {

	_, err = db.Exec("delete from forum_friend where user_id=? and create_by = ? ", forum.UserID, forum.CreateBy)
	return
}

// ForumFriendGet 查询单个关系
func ForumFriendGet(forum *moment.ForumFriend, db *sqlx.DB) error {
	return Get(forum, db)
}

type ForumFriendModel struct {
	ID              int     `db:"id"`
	UserID          int     `db:"user_id"`
	ForumID         int     `db:"forum_id"`
	Type            int     `db:"type"`
	ContentType     int     `db:"content_type"`
	Content         string  `db:"content"`
	Longitude       float64 `db:"longitude"`
	Latitude        float64 `db:"latitude"`
	LocationCity    string  `db:"location_city"`
	LocationName    string  `db:"location_name"`
	LocationAddress string  `db:"location_address"`
	CreateBy        int     `db:"create_by"`
	ContentEntity   string  `db:"content_entity"`
	Mention         string  `db:"mention"`
	Topic           string  `db:"topic"`
	Permission      int     `db:"permission"`
	PermissionUser  string  `db:"permission_user"`
	ThumbUp         int     `db:"thumb_up"`
	CommentCount    int     `db:"comment_count"`
	ViewState       int     `db:"view_state"`
}

func (f *ForumFriendModel) TableName() string {
	return "forum_friend"
}

//ForumFriendInsert 插入关系    bxh_debug
func ForumFriendInsert(forum *moment.ForumFriend, db *sqlx.DB, tx *sqlx.Tx) error {
	log.Printf("Topic:%v", string(forum.Topic))
	model := &ForumFriendModel{
		ID:              int(forum.ID),
		UserID:          int(forum.UserID),
		ForumID:         int(forum.ForumID),
		Type:            int(forum.Type),
		ContentType:     int(forum.ContentType),
		Content:         forum.Content,
		Longitude:       forum.Longitude,
		Latitude:        forum.Latitude,
		LocationCity:    forum.LocationCity,
		LocationName:    forum.LocationName,
		LocationAddress: forum.LocationAddress,
		CreateBy:        int(forum.ID),
		ContentEntity:   forum.ContentEntity,
		Mention:         forum.Mention,
		Topic:           string(forum.Topic),
		Permission:      int(forum.Permission),
		PermissionUser:  forum.PermissionUser,
		ThumbUp:         int(forum.ThumbUp),
		CommentCount:    int(forum.CommentCount),
		ViewState:       int(forum.ViewState),
	}
	return Insert(model, db, tx)
}

//ForumFriendsPage 我的朋友列表
func ForumFriendsPage(req *moment.ForumFriendPageReq, db *sqlx.DB) (forums []*moment.ForumFriend, err error) {
	var offsetSQL = ""
	if req.ForumID > 0 {
		offsetSQL = fmt.Sprintf(" and a.forum_id < %d ", req.ForumID)
	}
	sql := `
	select * from forum_friend a  where user_id = ?  ` + offsetSQL + `  
	ORDER BY a.forum_id DESC LIMIT ?
	`
	err = db.Select(&forums, sql, req.UserID, req.Limit)
	return
}

//ForumGetWithMedia 查询带媒体的帖子
func ForumGetWithMedia(req *moment.ForumGetWithMediaReq, db *sqlx.DB) (forums []*moment.ForumFriend, err error) {
	var offsetSQL = ""
	if req.ForumID > 0 {
		offsetSQL = fmt.Sprintf(" and a.forum_id < %d ", req.ForumID)
	}
	sql := `
	select * from forum_friend a  where user_id = ? and create_by = ? and content_type >1  ` + offsetSQL + `  
	ORDER BY a.forum_id DESC LIMIT ?
	`
	err = db.Select(&forums, sql, req.UserID, req.UserID, req.Limit)
	return
}

//ForumSelfMainPage 查看我自己的首页
func ForumSelfMainPage(req *moment.SelfMainPageReq, db *sqlx.DB) (forums []*moment.ForumFriend, err error) {
	var offsetSQL = ""
	if req.ForumID > 0 {
		offsetSQL = fmt.Sprintf(" and a.forum_id < %d", req.ForumID)
	}
	sql := `
	select * from forum_friend a  where  user_id = ? and  a.create_by  = ? ` + offsetSQL + `
	ORDER BY
	a.forum_id DESC 
	LIMIT ?
	`
	err = db.Select(&forums, sql, req.UserID, req.UserID, req.Limit)
	return
}

//ForumOtherMainPage 查看他人的首页
func ForumOtherMainPage(req *moment.OtherMainPageReq, db *sqlx.DB) (forums []*moment.ForumFriend, err error) {
	var offsetSQL = ""
	if req.ForumID > 0 {
		offsetSQL = fmt.Sprintf(" and a.forum_id < %d", req.ForumID)
	}
	sql := `
	select * from forum_friend a  where user_id = ? and a.create_by  = ?  and permission != 2` + offsetSQL + `
	ORDER BY
	a.forum_id DESC 
	LIMIT ?
	`
	err = db.Select(&forums, sql, req.UserID, req.FriendID, req.Limit)
	return
}

//ForumOtherMainByMouth 查看他人的首页
func ForumOtherMainByMouth(req *moment.ForumOtherMainByMouthReq, db *sqlx.DB) (forums []*moment.ForumFriend, err error) {
	forums = make([]*moment.ForumFriend, 0)
	sql := `
	select * from forum_friend a  where user_id = ? and a.create_by  = ? and a.create_at >=? and a.create_at <=? and permission != 2
	ORDER BY
	a.forum_id DESC 
	`
	err = db.Select(&forums, sql, req.UserID, req.FriendID, req.StartAt, req.EndAt)
	return
}

//ParticipatingFriends 帖子共同好友参与
func ParticipatingFriends(req *moment.ParticipatingFriendsMsg, db *sqlx.DB) ([]int32, error) {
	var inSQL string
	for _, firend := range req.Friends {
		inSQL += "," + strconv.Itoa(int(firend))
	}
	if inSQL != "" {
		inSQL = inSQL[1:]
	}
	var user []int32
	var err error
	var sql = `select a.create_by from thumb a where a.main_id = ? and  a.create_by in (` + inSQL + `)
	union 
	select b.create_by from comment b where b.forum_id = ? and b.create_by in (` + inSQL + `)`
	if err = db.Select(&user, sql, req.ForumID, req.ForumID); err != nil {
		return nil, err
	}
	return user, err

}

//ForumFriendDisableView 更新帖子可见权限
func ForumFriendDisableView(req *moment.ForumFriend, db *sqlx.DB) error {
	sql := "update forum_friend set view_state = ? where user_id = ? and create_by = ?"
	_, err := db.Exec(sql, req.ViewState, req.UserID, req.CreateBy)
	return err
}

//ForumCommentCountUpdate 更新帖子ID
func ForumCommentCountUpdate(froumID int64, commentCount int64, db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("update forum_topic set comment_count = comment_count+? where forum_id = ?", commentCount, froumID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("update forum_friend set comment_count = comment_count+? where forum_id = ?", commentCount, froumID); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("update forum_recommend set comment_count = comment_count+? where forum_id = ?", commentCount, froumID); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return err
}

//ForumThumbUpdate 更新帖子ID
func ForumThumbUpdate(froumID int64, thumbCount int64, db *sqlx.DB) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	if _, err = tx.Exec("update forum_topic set thumb_up = thumb_up+? where forum_id = ?", thumbCount, froumID); err != nil {
		tx.Rollback()
		return err
	}

	if _, err = tx.Exec("update forum_friend set thumb_up = thumb_up+? where forum_id = ?", thumbCount, froumID); err != nil {
		tx.Rollback()
		return err
	}
	if _, err = tx.Exec("update forum_recommend set thumb_up = thumb_up+? where forum_id = ?", thumbCount, froumID); err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return err
}
