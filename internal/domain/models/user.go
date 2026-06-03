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
	RegisterTime time.Time   `json:"registerTime"`
	FriendsReqs  []FriendReq `json:"friendsReqs"`
	Friends      []Friend    `json:"-"`
	BlockedUsers []string    `json:"blocked-users"`
}

type Friend struct {
	ID       uuid.UUID `json:"id"`
	Nickname string    `json:"nickname"`
}

type FriendReq struct {
	Nickname string    `json:"nickname"`
	ReqTime  time.Time `json:"reqTime"`
}
