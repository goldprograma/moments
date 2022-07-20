package domain

import "encoding/json"

//UserDomain 用户领域
type UserDomain struct {
	UserID     int             `json:"user_id"`
	AccessHash int             `json:"access_hash"`
	FirstName  string          `json:"first_name"`
	Sex        int             `json:"sex,omitempty"`
	Birthday   int             `json:"birthday,omitempty"`
	Photo      json.RawMessage `json:"photo,omitempty"`
}
