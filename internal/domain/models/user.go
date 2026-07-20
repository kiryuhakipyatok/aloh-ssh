package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	PersonalData PersonalData `json:"personalData"`
	Key          string       `json:"key"`
	Fingerprint  string       `json:"fingerprint"`
	Password     []byte       `json:"-"`
}

type PersonalData struct {
	Identity     Identity    `json:"identity"`
	Tagline      string      `json:"tagline"`
	Color        string      `json:"string"`
	RegisterTime time.Time   `json:"registerTime"`
	FriendsReqs  []FriendReq `json:"friendsReqs"`
	Friends      []Friend    `json:"friends"`
	BlockedUsers []Identity  `json:"blocked-users"`
}

type Identity struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
}

type Friend struct {
	Identity `json:"identity"`
	//FriendDenoises   `json:"denoises"`
	FriendAppereance `json:"appereance"`
}

type FriendAppereance struct {
	Tagline string `json:"tagline"`
	Color   string `json:"color"`
}

// type FriendDenoises struct {
// 	HardDenoised bool `json:"hard-denoised"`
// 	SoftDenoised bool `json:"soft-denoised"`
// }

type FriendReq struct {
	Identity `json:"identity"`
	ReqTime  time.Time `json:"reqTime"`
}
