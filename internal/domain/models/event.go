package models

const (
	NEW_FRIEND_REQ = iota
	ACCEPT_FRIEND
	DENY_FRIEND
	DELETE_FRIEND
	BLOCK_USER
	UNBLOCK_USER
)

type Event struct {
	Type uint   `json:"type"`
	Data string `json:"data"`
}

func NewFriendEvent(nickname string) Event {
	return Event{
		Type: NEW_FRIEND_REQ,
		Data: nickname,
	}
}

func AcceptFriendEvent(nickname string) Event {
	return Event{
		Type: ACCEPT_FRIEND,
		Data: nickname,
	}
}

func DenyFriendEvent(nickname string) Event {
	return Event{
		Type: DENY_FRIEND,
		Data: nickname,
	}
}

func DeleteFriendEvent(nickname string) Event {
	return Event{
		Type: DELETE_FRIEND,
		Data: nickname,
	}
}

func BlockUserEvent(nickname string) Event {
	return Event{
		Type: BLOCK_USER,
		Data: nickname,
	}
}

func UnblockUserEvent(nickname string) Event {
	return Event{
		Type: UNBLOCK_USER,
		Data: nickname,
	}
}
