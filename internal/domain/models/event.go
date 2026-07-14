package models

import (
	"encoding/json"

	"github.com/google/uuid"
)

const (
	NEW_FRIEND_REQ = iota
	ACCEPT_FRIEND
	DENY_FRIEND
	DELETE_FRIEND
	BLOCK_USER
	UNBLOCK_USER
	FRIEND_ONLINE
	FRIEND_OFFLINE
	FRIEND_CONNECTIONS
	UPDATE_HARD_DENOISE
	UPDATE_SOFT_DENOISE
	UPDATE_TAGLINE
)

type Event struct {
	Type uint            `json:"type"`
	Data json.RawMessage `json:"data"`
}

type FriendConnsData struct {
	Identity Identity `json:"identity"`
	Connects []string `json:"connects"`
}

type UsersHardDenoiseData struct {
	Identity Identity `json:"identity"`
	State    bool     `json:"state"`
}

type UsersSoftDenoiseData struct {
	Identity Identity `json:"identity"`
	State    bool     `json:"state"`
}

type TaglineData struct {
	Identity Identity `json:"identity"`
	Tagline  string   `json:"tagline"`
}

func NewFriendEvent(fp []byte) Event {
	return Event{
		Type: NEW_FRIEND_REQ,
		Data: fp,
	}
}

func AcceptFriendEvent(fp []byte) Event {
	return Event{
		Type: ACCEPT_FRIEND,
		Data: fp,
	}
}

func DenyFriendEvent(fp []byte) Event {
	return Event{
		Type: DENY_FRIEND,
		Data: fp,
	}
}

func DeleteFriendEvent(fp []byte) Event {
	return Event{
		Type: DELETE_FRIEND,
		Data: fp,
	}
}

func BlockUserEvent(fp []byte) Event {
	return Event{
		Type: BLOCK_USER,
		Data: fp,
	}
}

func UnblockUserEvent(fp []byte) Event {
	return Event{
		Type: UNBLOCK_USER,
		Data: fp,
	}
}

func FriendOnlineEvent(id []byte) Event {
	return Event{
		Type: FRIEND_ONLINE,
		Data: id,
	}
}

func FriendOfflineEvent(id []byte) Event {
	return Event{
		Type: FRIEND_OFFLINE,
		Data: id,
	}
}

func UpdateHardDenoiseEvent(uhdd []byte) Event {
	return Event{
		Type: UPDATE_HARD_DENOISE,
		Data: uhdd,
	}
}

func UpdateSoftDenoiseEvent(usdd []byte) Event {
	return Event{
		Type: UPDATE_SOFT_DENOISE,
		Data: usdd,
	}
}

func UpdateTaglineEvent(td []byte) Event {
	return Event{
		Type: UPDATE_TAGLINE,
		Data: td,
	}
}

func MarshIdentity(id uuid.UUID, nickname string) ([]byte, error) {
	iden := Identity{
		ID:       id,
		Nickname: nickname,
	}
	data, err := json.Marshal(iden)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func MarshTD(id uuid.UUID, nickname string, tagline string) ([]byte, error) {
	td := TaglineData{
		Identity: Identity{
			ID:       id,
			Nickname: nickname,
		},
	}
	data, err := json.Marshal(td)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func MarshID(id uuid.UUID) ([]byte, error) {
	data, err := json.Marshal(id)
	if err != nil {
		return nil, err
	}

	return data, nil
}
