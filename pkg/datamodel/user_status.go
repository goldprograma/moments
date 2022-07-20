package datamodel

// UserStatus model
type UserStatus struct {
	ID          int `db:"id" json:"ID"`
	UserID      int `db:"user_id" json:"UserID,omitempty"`
	ForumCount  int `db:"forum_count" json:"ForumCount,omitempty"`
	FollowCount int `db:"follow_count" json:"FollowCount,omitempty"`
	ThumbCount  int `db:"thumb_count" json:"ThumbCount,omitempty"`
	FansCount   int `db:"fans_count" json:"FansCount,omitempty"`
	//?遗留字段：好友最新动态ForumID
	FriendVersion string `db:"friend_version" json:"FriendVersion,omitempty"`
	//?遗留字段：关注的人最新动态ForumID
	FollowVersion string `db:"follow_version" json:"FollowVersion,omitempty"`
	//?最后已读系统推荐动态ForumID
	RecommendVersionRead string `db:"recommend_version_read" json:"RecommendVersionRead,omitempty"`
	//?最后已读好友动态ForumID
	FriendVersionRead string `db:"friend_version_read" json:"FriendVersionRead,omitempty"`
	//?最后已读关注的人动态ForumID
	FollowVersionRead string `db:"follow_version_read" json:"FollowVersionRead,omitempty"`
	HomeBackground    string `db:"home_background" json:"HomeBackground,omitempty"`
	//time Unix
	CreateAt int `db:"create_at" json:"CreateAt,omitempty"`
	//time Unix
	UpdateAt int `db:"update_at" json:"UpdateAt,omitempty"`
}
