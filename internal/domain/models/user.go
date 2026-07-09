package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID    `json:"id"`
	PersonalData PersonalData `json:"personalData"`
	Key          string       `json:"key"`
	Fingerprint  string       `json:"fingerprint"`
	Password     []byte       `json:"-"`
}

type PersonalData struct {
	Nickname     string      `json:"nickname"`
	Tagline      string      `json:"tagline"`
	RegisterTime time.Time   `json:"registerTime"`
	FriendsReqs  []FriendReq `json:"friendsReqs"`
	Friends      []Friend    `json:"friends"`
	BlockedUsers []string    `json:"blocked-users"`
}

type Friend struct {
	FriendPersonal `json:"personal"`
	//FriendDenoises   `json:"denoises"`
	FriendAppereance `json:"appereance"`
}

type FriendAppereance struct {
	Tagline string `json:"tagline"`
	//Color   string `json:"color"`
}

// type FriendDenoises struct {
// 	HardDenoised bool `json:"hard-denoised"`
// 	SoftDenoised bool `json:"soft-denoised"`
// }

type FriendPersonal struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
}

type FriendReq struct {
	FriendPersonal
	ReqTime time.Time `json:"reqTime"`
}
