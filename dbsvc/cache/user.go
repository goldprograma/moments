package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"gitlab.moments.im/dbsvc/internal"
	"gitlab.moments.im/pkg/datamodel"
	"gitlab.moments.im/pkg/protoc/moment"

	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

var (
	UserStatisticsKey       = "moments:user:%d:statistics"
	UserVersionKey          = "moments:user:%d:version"
	UserRecommendVersionKey = "moments:user:recommendversion"
	UserRecommendKey        = "moments:user:recommend"
	UserForumsKey           = "moments:user:%d:forum"
	UserAlbumKey            = "moments:user:%d:album"

	hash_field_forum_count    = "ForumCount"
	hash_field_thumb_count    = "ThumbCount"
	hash_field_homebackground = "HomeBackground"
)

//UserStatisticsCache 用户统计
// func UserStatisticsCache(cc *redis.ClusterClient, db *sqlx.DB) (err error) {
// 	// if err = internal.UserStatusUpdate(db); err != nil {
// 	// 	return err
// 	// }
// 	var userStates []*moment.UserStatus
// 	if userStates, err = internal.UserStatusAll(db); err != nil {
// 		return err
// 	}

// 	for _, userState := range userStates {
// 		//if err = cc.HSet(fmt.Sprintf(UserStatisticsKey, userState.UserID),
// 		//	"FansCount", userState.FansCount, "ForumCount", userState.ForumCount,
// 		//	"FollowCount", userState.FollowCount, "ThumbCount", userState.ThumbCount).Err(); err != nil {
// 		//	return err
// 		//}

// 		_, err = cc.Pipelined(func(p redis.Pipeliner) error {

// 			var (
// 				err error
// 				key string
// 			)
// 			key = fmt.Sprintf(UserStatisticsKey, userState.UserID)
// 			if err = p.HSet(key, "FansCount", userState.FansCount).Err(); err != nil {
// 				fmt.Println("UserStatisticsCache error ", err)
// 			}
// 			if err = p.HSet(key, "ForumCount", userState.ForumCount).Err(); err != nil {
// 				fmt.Println("UserStatisticsCache error ", err)
// 			}
// 			if err = p.HSet(key, "FollowCount", userState.FollowCount).Err(); err != nil {
// 				fmt.Println("UserStatisticsCache error ", err)
// 			}
// 			if err = p.HSet(key, "ThumbCount", userState.ThumbCount).Err(); err != nil {
// 				fmt.Println("UserStatisticsCache error ", err)
// 			}
// 			return nil
// 		})

// 		//设置随机过期时间
// 		rd := rand.New(rand.NewSource(time.Now().UnixNano()))
// 		err = cc.Expire(fmt.Sprintf(UserStatisticsKey, userState.UserID), time.Second*time.Duration(rd.Int63n(60))).Err()
// 	}
// 	return err
// }

//UserForumCacheSet 缓存用户帖子
func UserForumCacheSet(friend *moment.ForumFriend, redis *redis.ClusterClient) (err error) {
	data, _ := json.Marshal(friend)
	err = redis.Set(fmt.Sprintf(UserForumsKey, friend.CreateBy), data, 0).Err()
	return err
}

//UserAlbumCacheSet 缓存用户帖子
func UserAlbumCacheSet(userID int32, medias []*moment.Media, redis *redis.ClusterClient) (err error) {
	for _, media := range medias {
		data, _ := json.Marshal(media)
		if err = redis.LPush(fmt.Sprintf(UserAlbumKey, userID), data).Err(); err != nil {
			return err
		}
	}
	redis.LTrim(fmt.Sprintf(UserAlbumKey, userID), 0, 4)
	return err
}

//UserAlbumCacheGet 缓存用户相册
func UserAlbumCacheGet(userID int32, medias []*moment.Media, redis *redis.ClusterClient, db *sqlx.DB) (err error) {
	err = redis.LRange(fmt.Sprintf(UserAlbumKey, userID), 0, 4).ScanSlice(medias)
	if len(medias) == 0 {
		var forumIDs []int64
		if err = db.Select(&forumIDs, "select forum_id from forum_friend where user_id =? and create_by = ? and permission=1 and content_type>1 order by forum_id desc limit 4", userID, userID); err != nil {
			return err
		}
		for _, forumID := range forumIDs {
			if len(medias) >= 4 {
				break
			}
			var media []*moment.Media
			if err = db.Select(&media, "select * from media where main_id = ?", forumID); err != nil {
				return
			}
			medias = append(medias, media...)
		}
		go UserAlbumCacheSet(userID, medias, redis)
	}
	return err
}

//UserForumCacheGet 用户帖子
func UserForumCacheGet(friendForum *moment.ForumFriend, redisCli *redis.ClusterClient, db *sqlx.DB) error {

	data, err := redisCli.Get(fmt.Sprintf(UserForumsKey, friendForum.CreateBy)).Bytes()
	if err != nil {
		if err == redis.Nil {
			if err = db.Get(friendForum, "select * from forum_friend where user_id= ? and create_by = ? order by forum_id limit 1", friendForum.CreateBy, friendForum.CreateBy); err != nil {
				return err
			}
			if err = UserForumCacheSet(friendForum, redisCli); err != nil {
				log.Println("缓存用户最新帖子失败", friendForum.CreateBy)
			}
			return nil
		}
		return err

	}
	err = json.Unmarshal(data, friendForum)

	return err
}

//asyncCacheUpdateUserHomeBackground 异步-更新用户统计缓存中的背景图
func AsyncCacheUpdateUserHomeBackground(redisClient *redis.ClusterClient, user_id int, homebackground string, logger *zap.Logger) {
	key_user_status_homebg := fmt.Sprintf(UserStatisticsKey, user_id)
	isExist, err := redisClient.HExists(key_user_status_homebg, hash_field_homebackground).Result()
	if err != nil {
		logger.Error("asyncCacheUpdateUserHomeBackground HExists => " + err.Error())
		return
	}
	if !isExist {
		return
	}
	err = redisClient.HSet(key_user_status_homebg, hash_field_homebackground, homebackground).Err()
	if err != nil {
		logger.Error("asyncCacheUpdateUserHomeBackground HSet => " + err.Error())
	}
}

//UserStatusUpdateCache 更新用户统计缓存
func UserStatusUpdateCache(userState *moment.UserStatus, redis *redis.ClusterClient) (err error) {
	var has bool
	if userState.ForumCount != 0 {
		if has, err = redis.HExists(fmt.Sprintf(UserStatisticsKey, userState.UserID), "ForumCount").Result(); err != nil {
			return err
		}
		if has {
			if err = redis.HIncrBy(fmt.Sprintf(UserStatisticsKey, userState.UserID), "ForumCount", userState.ForumCount).Err(); err != nil {
				return
			}
		}
	}
	if userState.FollowCount != 0 {
		if has, err = redis.HExists(fmt.Sprintf(UserStatisticsKey, userState.UserID), "FollowCount").Result(); err != nil {
			return err
		}
		if has {
			if err = redis.HIncrBy(fmt.Sprintf(UserStatisticsKey, userState.UserID), "FollowCount", userState.FollowCount).Err(); err != nil {
				return
			}
		}
	}
	if userState.ThumbCount != 0 {
		if has, err = redis.HExists(fmt.Sprintf(UserStatisticsKey, userState.UserID), "ThumbCount").Result(); err != nil {
			return err
		}
		if has {
			if err = redis.HIncrBy(fmt.Sprintf(UserStatisticsKey, userState.UserID), "ThumbCount", userState.ThumbCount).Err(); err != nil {
				return
			}
		}
	}
	if userState.FansCount != 0 {
		if has, err = redis.HExists(fmt.Sprintf(UserStatisticsKey, userState.UserID), "FansCount").Result(); err != nil {
			return err
		}
		if has {
			if err = redis.HIncrBy(fmt.Sprintf(UserStatisticsKey, userState.UserID), "FansCount", userState.FansCount).Err(); err != nil {
				return
			}
		}
	}
	// //设置随机过期时间
	// rd := rand.New(rand.NewSource(time.Now().UnixNano()))
	// err = redis.Expire(fmt.Sprintf(UserStatisticsKey, userState.UserID), time.Second*time.Duration(rd.Int63n(60))).Err()

	return err
}

// UserStatisticsCacheGet 获取用户朋友圈统计数据
// func UserStatisticsCacheGet(userState *moment.UserStatus, loger *zap.Logger, cc *redis.ClusterClient, db *sqlx.DB) (err error) {
// 	data, err := cc.HGetAll(fmt.Sprintf(UserStatisticsKey, userState.UserID)).Result()
// 	if err != nil {
// 		if err != redis.Nil {
// 			return err
// 		}
// 	}
// 	if len(data) == 0 {
// 		// 查询数据库 放入		var userStatisticsRep = &moment.UserStatisticsRep{}
// 		err = internal.UserStatistics(userState, db)
// 		if err != nil {
// 			return err
// 		}
// 		// 更新用户
// 		go func() {
// 			_, err = cc.Pipelined(func(p redis.Pipeliner) error {
// 				var (
// 					err error
// 					key string
// 				)
// 				key = fmt.Sprintf(UserStatisticsKey, userState.UserID)
// 				if err = p.HSet(key, "FansCount", userState.FansCount).Err(); err != nil {
// 					loger.Error("更新用户统计(FansCount)失败", zap.Int32("用户", userState.UserID), zap.Error(err))
// 				}
// 				if err = p.HSet(key, "ForumCount", userState.ForumCount).Err(); err != nil {
// 					loger.Error("更新用户统计(ForumCount)失败", zap.Int32("用户", userState.UserID), zap.Error(err))
// 				}
// 				if err = p.HSet(key, "FollowCount", userState.FollowCount).Err(); err != nil {
// 					loger.Error("更新用户统计(FollowCount)失败", zap.Int32("用户", userState.UserID), zap.Error(err))
// 				}
// 				if err = p.HSet(key, "ThumbCount", userState.ThumbCount).Err(); err != nil {
// 					loger.Error("更新用户统计(ThumbCount)失败 ", zap.Int32("用户", userState.UserID), zap.Error(err))
// 				}
// 				return nil
// 			})

// 			//设置随机过期时间
// 			rd := rand.New(rand.NewSource(time.Now().UnixNano()))
// 			if err = cc.Expire(fmt.Sprintf(UserStatisticsKey, userState.UserID), time.Second*time.Duration(rd.Int63n(60))).Err(); err != nil {
// 				loger.Error("设置缓存过期时间失败", zap.Int32("用户", userState.UserID), zap.Error(err))
// 			}
// 		}()

// 		return nil
// 	}

// 	if data["FansCount"] != "" {
// 		userState.FansCount, err = strconv.ParseInt(data["FansCount"], 0, 64)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if data["FollowCount"] != "" {
// 		userState.FollowCount, err = strconv.ParseInt(data["FollowCount"], 0, 64)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if data["ForumCount"] != "" {
// 		userState.ForumCount, err = strconv.ParseInt(data["ForumCount"], 0, 64)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if data["ThumbCount"] != "" {
// 		userState.ThumbCount, err = strconv.ParseInt(data["ThumbCount"], 0, 64)
// 		if err != nil {
// 			return err
// 		}
// 	}
// 	if data["HomeBackground"] != "" {
// 		userState.HomeBackground = data["HomeBackground"]
// 	}

// 	return nil
// }

//GetUserStatusByUserID 获取用户朋友圈统计数据
func GetUserStatusByUserID(mysqlClient *sqlx.DB, redisClient *redis.ClusterClient, logger *zap.Logger, user_id int) (userStatus moment.UserStatus, err error) {
	redis_key := fmt.Sprintf(UserStatisticsKey, user_id)
	redisIsExist := true
	redisCorrect := true

	hashmap, err := redisClient.HGetAll(redis_key).Result()
	if err != nil {
		redisIsExist = false
		//key不存在
		if err != redis.Nil {
			logger.Error("GetUserStatusByUserID redisClient.HGetAll => " + err.Error())
		}
	}

	//缓存key存在
	if redisIsExist {
		homebackground := hashmap[hash_field_homebackground]
		forumCount, errForumCount := strconv.Atoi(hashmap[hash_field_forum_count])
		thumbCount, errThumbCount := strconv.Atoi(hashmap[hash_field_thumb_count])
		if errForumCount != nil || errThumbCount != nil {
			redisCorrect = false
			//数据不正确，删除key
			redisClient.Del(redis_key)
		}

		//缓存数据正确
		if redisCorrect {
			userStatus = moment.UserStatus{
				UserID:         int32(user_id),
				HomeBackground: homebackground,
				ForumCount:     int64(forumCount),
				ThumbCount:     int64(thumbCount),
			}
			return userStatus, nil
		}
	}

	//缓存key不存在或数据不正确，从mysql查询
	userStatusModel, err := internal.GetUserStatusByUserID(mysqlClient, user_id)
	if err != nil {
		logger.Error("GetUserStatusByUserID internal.UserStatus => " + err.Error())
		return userStatus, err
	}
	userStatus = moment.UserStatus{
		UserID:         int32(user_id),
		HomeBackground: userStatusModel.HomeBackground,
		ForumCount:     int64(userStatusModel.ForumCount),
		ThumbCount:     int64(userStatusModel.ThumbCount),
	}
	//查询结果存入redis
	go func() {
		_, err = redisClient.Pipelined(func(pi redis.Pipeliner) error {
			err = pi.HSet(redis_key, hash_field_forum_count, userStatus.ForumCount).Err()
			if err != nil {
				return err
			}
			err = pi.HSet(redis_key, hash_field_thumb_count, userStatus.ThumbCount).Err()
			if err != nil {
				return err
			}
			err = pi.HSet(redis_key, hash_field_homebackground, userStatus.HomeBackground).Err()
			return nil
		})
		if err != nil {
			logger.Error("GetUserStatusByUserID redisClient.Pipelined => " + err.Error())
		} else {
			redisClient.Expire(redis_key, time.Hour*24*7)
		}
	}()

	return userStatus, nil
}

// UserStatusVersionGet 用户单个版本获取
func UserStatusVersionGet(userState *moment.UserStatus, cc *redis.ClusterClient, db *sqlx.DB) (err error) {

	//判断用户是否存在
	if cc.Exists(fmt.Sprintf(UserVersionKey, userState.UserID)).Val() == 0 { //不存在
		if err = internal.UserVersion(userState, db); err != nil {
			return err
		}
		if userState.ID == 0 { //没有用户信息则创建
			userState.CreateAt = time.Now().Unix()
			// err = internal.Insert(userState, db, nil)
			create_at := time.Now().Unix()
			newUserStatus := datamodel.UserStatus{
				UserID:   int(userState.UserID),
				CreateAt: int(create_at),
				UpdateAt: int(create_at),
			}
			err = internal.InsertUserStatus(newUserStatus, db)
			if err != nil {
				return err
			}

		}
		//err = cc.HSet(fmt.Sprintf(UserVersionKey, userState.UserID),
		//	"RecommendVersion", userState.RecommendVersion,
		//	"RecommendUID", userState.RecommendUID,
		//	"FollowVersion", userState.FollowVersion,
		//	"FollowUID", userState.FollowUID,
		//	"FriendVersion", userState.FriendVersion,
		//	"FriendUID", userState.FriendUID,
		//	"RecommendVersionRead", userState.RecommendVersionRead,
		//	"FriendVersionRead", userState.FriendVersionRead,
		//	"FollowVersionRead", userState.FollowVersionRead,
		//).Err()
		_, err = cc.Pipelined(func(p redis.Pipeliner) error {
			key := fmt.Sprintf(UserVersionKey, userState.UserID)
			p.HSet(key, "RecommendVersion", userState.RecommendVersion)
			p.HSet(key, "RecommendUID", userState.RecommendUID)
			p.HSet(key, "FollowVersion", userState.FollowVersion)
			p.HSet(key, "FollowUID", userState.FollowUID)
			p.HSet(key, "FriendVersion", userState.FriendVersion)
			p.HSet(key, "FriendUID", userState.FriendUID)
			p.HSet(key, "RecommendVersionRead", userState.RecommendVersionRead)
			p.HSet(key, "FriendVersionRead", userState.FriendVersionRead)
			p.HSet(key, "FollowVersionRead", userState.FollowVersionRead)
			return nil
		})
		return
	}

	if userState.RecommendVersion, err = cc.Get(UserRecommendVersionKey).Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}

	var recommendUID int64
	if recommendUID, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "RecommendUID").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	userState.RecommendUID = int32(recommendUID)

	if userState.FriendVersion, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "FriendVersion").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	var friendUID int64
	if friendUID, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "FriendUID").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	userState.FriendUID = int32(friendUID)

	if userState.FollowVersion, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "FollowVersion").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	var followUID int64
	if followUID, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "FollowUID").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	userState.FollowUID = int32(followUID)

	if userState.RecommendVersionRead, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "RecommendVersionRead").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	if userState.FriendVersionRead, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "FriendVersionRead").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	if userState.FollowVersionRead, err = cc.HGet(fmt.Sprintf(UserVersionKey, userState.UserID), "FollowVersionRead").Int64(); err != nil {
		if err != redis.Nil {
			return
		}
	}
	return nil
}

//InitUserStatus 初始化用户user_status表
func InitUserStatus(db *sqlx.DB, user_id int, logger *zap.Logger) error {
	_, isExist, err := internal.GetUserStatus(db, user_id, logger)
	if err != nil {
		logger.Error("InitUserStatus internal.GetUserStatus => " + err.Error())
		return err
	}
	if isExist {
		return nil
	}

	create_at := time.Now().Unix()
	newUserStatus := datamodel.UserStatus{
		UserID:   user_id,
		CreateAt: int(create_at),
		UpdateAt: int(create_at),
	}
	err = internal.InsertUserStatus(newUserStatus, db)
	if err != nil {
		logger.Error("InitUserStatus internal.InsertUserStatus => " + err.Error())
	}
	return err
}

// UserStatusVersionReadUpdate 用户单个版本获取
func UserStatusVersionReadUpdate(userState *moment.UserStatus, redisCli *redis.ClusterClient, db *sqlx.DB) (err error) {
	// if userState.RecommendVersionRead != "" || userState.FriendVersionRead != "" || userState.FollowVersionRead != "" {
	// if _, err = db.Exec("update user_status set recommend_version_read=?,friend_version_read=?,follow_version_read=? where user_id = ?",
	// 	userState.RecommendVersionRead, userState.FriendVersionRead, userState.FollowVersionRead, userState.UserID); err != nil {
	// 	return err
	// }
	if userState.RecommendVersionRead != 0 {
		if err = redisCli.HSet(fmt.Sprintf(UserVersionKey, userState.UserID), "RecommendVersionRead", userState.RecommendVersionRead).Err(); err != nil {
			return
		}
	} else {
		if userState.RecommendVersion, err = redisCli.Get(UserRecommendVersionKey).Int64(); err != nil {
			if err != redis.Nil {
				return
			}
		}
		if err = redisCli.HSet(fmt.Sprintf(UserVersionKey, userState.UserID), "RecommendVersionRead", userState.RecommendVersion).Err(); err != nil {
			return
		}
	}

	if userState.FriendVersionRead != 0 {
		if err = redisCli.HSet(fmt.Sprintf(UserVersionKey, userState.UserID), "FriendVersionRead", userState.FriendVersionRead).Err(); err != nil {
			return
		}
	} else {
		//if err = redisCli.HSet(fmt.Sprintf(UserVersionKey, userState.UserID), "FriendVersionRead", 0, "FriendVersion", 0).Err(); err != nil {
		//	return
		//}
		_, err = redisCli.Pipelined(func(p redis.Pipeliner) error {
			key := fmt.Sprintf(UserVersionKey, userState.UserID)
			if err := p.HSet(key, "FriendVersionRead", 0).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "FriendVersion", 0).Err(); err != nil {
				return err
			}
			return nil
		})
	}

	if userState.FollowVersionRead != 0 {
		if err = redisCli.HSet(fmt.Sprintf(UserVersionKey, userState.UserID), "FollowVersionRead", userState.FollowVersionRead).Err(); err != nil {
			return
		}
	} else {
		//if err = redisCli.HSet(fmt.Sprintf(UserVersionKey, userState.UserID), "FollowVersion", 0, "FollowVersionRead", 0).Err(); err != nil {
		//	return
		//}
		_, err = redisCli.Pipelined(func(p redis.Pipeliner) error {
			key := fmt.Sprintf(UserVersionKey, userState.UserID)
			if err := p.HSet(key, "FollowVersion", 0).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "FollowVersionRead", 0).Err(); err != nil {
				return err
			}
			return nil
		})
	}

	// }
	return nil
}

// UserStatusVersionUpdate 缓存
func UserStatusVersionUpdate(req *moment.UserVersionUpdateReq, cc *redis.ClusterClient, db *sqlx.DB) (err error) {

	if req.IsRecommedUser { // 是推荐用户
		// if _, err = db.Exec("update user_recommend_version set version = ?", req.Version); err != nil {
		// 	return
		// }

		if err = cc.Set(UserRecommendVersionKey, req.Version, 0).Err(); err != nil {
			return err
		}
		//if err = cc.HSet(fmt.Sprintf(UserVersionKey, req.UserID), "RecommendVersionRead", req.Version, "RecommendUID", req.UserID).Err(); err != nil {
		//	return err
		//}
		_, err = cc.Pipelined(func(p redis.Pipeliner) error {
			key := fmt.Sprintf(UserVersionKey, req.UserID)
			if err := p.HSet(key, "RecommendVersionRead", req.Version).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "RecommendUID", req.UserID).Err(); err != nil {
				return err
			}
			return nil
		})

	}

	for _, friend := range req.Friend {
		//if err = cc.HSet(fmt.Sprintf(UserVersionKey, friend), "FriendVersion", req.Version, "FriendUID", req.UserID, "UpdateAt", time.Now()).Err(); err != nil {
		//	return err
		//}
		_, err = cc.Pipelined(func(p redis.Pipeliner) error {
			key := fmt.Sprintf(UserVersionKey, friend)
			if err := p.HSet(key, "FollowVersion", req.Version).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "FollowUID", req.UserID).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "UpdateAt", time.Now()).Err(); err != nil {
				return err
			}
			return nil
		})
	}

	//所有关注我的人

	for _, follow := range req.Fans {
		//if err = cc.HSet(fmt.Sprintf(UserVersionKey, follow), "FollowVersion", req.Version, "FollowUID", req.UserID, "UpdateAt", time.Now()).Err(); err != nil {
		//	return err
		//}
		_, err = cc.Pipelined(func(p redis.Pipeliner) error {
			key := fmt.Sprintf(UserVersionKey, follow)
			if err := p.HSet(key, "FollowVersion", req.Version).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "FollowUID", req.UserID).Err(); err != nil {
				return err
			}
			if err := p.HSet(key, "UpdateAt", time.Now()).Err(); err != nil {
				return err
			}
			return nil
		})
	}
	return err
}

//UserRecommendCache 推荐用户
func UserRecommendCache(cc *redis.ClusterClient, db *sqlx.DB) error {

	users, err := internal.RecommendAll(db)
	if err != nil {
		return err
	}
	// 清空redis 从DB 重新获取
	if err = cc.Del(UserRecommendKey).Err(); err != nil {
		return err
	}

	for _, user := range users {
		data, _ := json.Marshal(user)
		if err = cc.HSet(UserRecommendKey, strconv.Itoa(int(user.UserID)), data).Err(); err != nil {
			return err
		}
	}

	// log.Panic("---------------------------", err)

	return err
}

//UserRecommendCacheWatch 推荐用户监听
func UserRecommendCacheWatch(redis *redis.ClusterClient, db *sqlx.DB, log *zap.Logger) {
	//防止redis挂
	var (
		val int64
		err error
	)
	for {
		t := time.Tick(time.Second * 10)
		select {
		case <-t:
			//检查redis key 是否存在
			if val, err = redis.Exists(UserRecommendKey).Result(); err != nil {
				log.Error("检查推荐用户缓存失败", zap.Error(err))
			}
			//不存在 则缓存
			if val == 0 {
				if err = UserRecommendCache(redis, db); err != nil {
					log.Error("推荐用户 DB->Redis 缓存失败,10S后重试", zap.Error(err))
				}
			}
		}
	}

}

//UserRecommendGet 推荐用户
func UserRecommendGet(req *moment.UserRecommend, redisCli *redis.ClusterClient) error {
	data, err := redisCli.HGet(UserRecommendKey, strconv.Itoa(int(req.UserID))).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}
	if err = json.Unmarshal(data, req); err != nil {
		return err
	}
	return nil
}
