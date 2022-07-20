package internal

import (
	"database/sql"
	"fmt"

	"gitlab.moments.im/pkg/datamodel"
	"gitlab.moments.im/pkg/protoc/moment"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
)

//UserVersion 版本号
func UserVersion(userState *moment.UserStatus, db *sqlx.DB) (err error) {
	sql := "select * from user_status where user_id = ?"
	err = db.Get(userState, sql, userState.UserID)
	return
}

//UserStatus 状态
func UserStatus(userState *moment.UserStatus, db *sqlx.DB) (err error) {
	sql := "select * from user_status where user_id = ?"
	return db.Get(userState, sql, userState.UserID)
}

//GetUserStatusByUserID 通过user_id获取用户统计数据
func GetUserStatusByUserID(db *sqlx.DB, user_id int) (userStatus datamodel.UserStatus, err error) {
	sql_select_user_status := `select * from user_status where user_id = ?`
	err = db.Get(&userStatus, sql_select_user_status, user_id)
	return userStatus, err
}

//UserStatus 版本号
func UserStatusAll(db *sqlx.DB) (userState []*moment.UserStatus, err error) {
	sql := "select * from user_status"
	return userState, db.Select(&userState, sql)
}

//UserStatusUpdate 更新用户统计
// func UserStatusUpdate(db *sqlx.DB) (err error) {
// 	sql := `update user_status a
// 	set forum_count =   (select count(1) from forum b where b.create_by = a.user_id and b.status =1),
// 	 follow_count =  (select count(1) from follow c where a.user_id = c.create_by ),
// 	 fans_count =  (select count(1) from follow d where a.user_id = d.follow_uid ),
// 	thumb_count =  (select count(1) from thumb e  where a.user_id = e.forum_uid or a.user_id = e.comment_uid  )`
// 	_, err = db.Exec(sql)
// 	return
// }

//UserStatusUpdateDB 更新首页更新
func UserStatusUpdateDB(req *moment.UserStatus, db *sqlx.DB, tx *sqlx.Tx) (err error) {
	return Update(req, db, tx, "user_id")
}

//UserStatusThumbCountUpdateDB 更新点赞
func UserStatusThumbCountUpdateDB(req *moment.UserStatus, db *sqlx.DB) (err error) {
	_, err = db.Exec("update user_status set thumb_count = thumb_count + ? where user_id = ?", req.ThumbCount, req.UserID)
	return
}

//UserStatusFollowCountUpdateDB 更新点赞
func UserStatusFollowCountUpdateDB(req *moment.UserStatus, db *sqlx.DB) (err error) {
	_, err = db.Exec("update user_status set follow_count = follow_count + ? where user_id = ?", req.FollowCount, req.UserID)
	return
}

//UserStatusFansCountUpdateDB 更新点赞
func UserStatusFansCountUpdateDB(req *moment.UserStatus, db *sqlx.DB) (err error) {
	_, err = db.Exec("update user_status set fans_count = fans_count + ? where user_id = ?", req.FansCount, req.UserID)
	return
}

//UserStatusForumCountUpdateDB 更新发帖数
func UserStatusForumCountUpdateDB(req *moment.UserStatus, db *sqlx.DB) (err error) {
	_, err = db.Exec("update user_status set  forum_count = forum_count +? where user_id=? ", req.ForumCount, req.UserID)
	return err
}

// UserStatistics 用户统计
func UserStatistics(req *moment.UserStatus, db *sqlx.DB) error {
	var err error
	sql := `select * from user_status where user_id = ?;`
	// err = db.Get(req, sql, req.UserID)
	user := datamodel.UserStatus{}
	err = db.Get(&user, sql, req.UserID)
	if err != nil {
		return err
	}
	req.ForumCount = int64(user.ForumCount)
	req.ThumbCount = int64(user.ThumbCount)
	req.FollowCount = int64(user.FollowCount)
	req.FansCount = int64(user.FansCount)
	req.HomeBackground = user.HomeBackground
	req.CreateAt = int64(user.CreateAt)
	req.UpdateAt = int64(user.UpdateAt)
	return nil
}

// UserStatisticsUpdate 更新用户统计
func UserStatisticsUpdate(UserStatus *moment.UserStatus, db *sqlx.DB, tx *sqlx.Tx) error {
	return Update(UserStatus, db, tx, "user_id")
}

func UserHomeBackgroudGet(req *moment.UserStatus, db *sqlx.DB) error {
	return db.Get(&req.HomeBackground, "select IFNULL(home_background, '') from user_status where user_id = ?", req.UserID)
}

//UserAllID 所有用户ID
func UserAllID(req *moment.UserAllIDReq, db *sqlx.DB) (userids []int32, err error) {
	var (
		excludeSQL = ""
		notin      = ""
	)
	for _, uid := range req.ExcludeUID {
		notin += fmt.Sprintf(",%d", uid)
	}
	if notin != "" {
		excludeSQL = " where user_id not in(" + notin[1:] + ")"
	}

	err = db.Select(&userids, "select user_id from user_status "+excludeSQL)
	return
}

//GetUserStatus 根据user_id获取user_status
func GetUserStatus(db *sqlx.DB, user_id int, logger *zap.Logger) (userStatusModel datamodel.UserStatus, isExist bool, err error) {
	sql_select_user_status := "select * from user_status where user_id = ?"
	err = db.Get(&userStatusModel, sql_select_user_status, user_id)
	if err != nil {
		if err == sql.ErrNoRows {
			return userStatusModel, isExist, nil
		}
		logger.Error("GetUserStatus => " + err.Error())
		return userStatusModel, isExist, err
	}
	isExist = true
	return userStatusModel, isExist, err
}

//InsertUserStatus 初始化插入用户朋友圈状态表
func InsertUserStatus(user_status datamodel.UserStatus, db *sqlx.DB) error {
	sql_insert_user_status := `insert into user_status (
								user_id,
								forum_count,follow_count,thumb_count,fans_count,
								friend_version,follow_version,
								recommend_version_read,friend_version_read,follow_version_read,
								home_background,create_at,update_at)
								values (?,0,0,0,0,
									'','',
									'','','',
									'',?,?);`
	_, err := db.Exec(sql_insert_user_status, user_status.UserID, user_status.CreateAt, user_status.UpdateAt)
	return err
}
