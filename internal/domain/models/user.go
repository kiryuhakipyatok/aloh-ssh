package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	PersonalData PersonalData `json:"personalData"`
	Key          string
	Fingerprint  string
	Password     []byte
}

type PersonalData struct {
	Nickname     string      `json:"nickname"`
	RegisterTime time.Time   `json:"registerTime"`
	FriendsReqs  []FriendReq `json:"friendsReqs"`
	Friends      uuid.UUIDs  `json:"friends"`
	BlockedUsers uuid.UUIDs  `json:"blocked-users"`
}

type FriendReq struct {
	Nickname string    `json:"nickname"`
	ReqTime  time.Time `json:"reqTime"`
}
