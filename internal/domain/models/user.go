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
	Friends      []string    `json:"friends"`
	BlockedUsers []string    `json:"blocked-users"`
}

type FriendReq struct {
	Nickname string    `json:"nickname"`
	ReqTime  time.Time `json:"reqTime"`
}
