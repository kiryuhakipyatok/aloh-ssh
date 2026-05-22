package models

import (
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
	RegisterTime string      `json:"registerTime"`
	FriendsReqs  []FriendReq `json:"friendsReqs"`
	Friends      []string    `json:"friends"`
}

type FriendReq struct {
	Nickname string `json:"nickname"`
	ReqTime  string `json:"reqTime"`
}
