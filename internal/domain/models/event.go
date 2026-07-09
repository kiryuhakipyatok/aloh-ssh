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
	Id       uuid.UUID   `json:"id"`
	Connects []uuid.UUID `json:"connects"`
}

type UsersHardDenoiseData struct {
	Id    uuid.UUID `json:"id"`
	State bool      `json:"state"`
}

type UsersSoftDenoiseData struct {
	Id    uuid.UUID `json:"id"`
	State bool      `json:"state"`
}

type TaglineData struct {
	Id      uuid.UUID `json:"id"`
	Tagline string    `json:"tagline"`
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

func MarshFP(id uuid.UUID, nickname string) ([]byte, error) {
	fp := FriendPersonal{
		ID:       id,
		Nickname: nickname,
	}
	data, err := json.Marshal(fp)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func MarshTD(id uuid.UUID, tagline string) ([]byte, error) {
	td := TaglineData{
		Id:      id,
		Tagline: tagline,
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
