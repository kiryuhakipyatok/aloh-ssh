package models

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
)

type Event struct {
	Type uint   `json:"type"`
	Data []byte `json:"data"`
}

type FriendConnsData struct {
	Nickname string   `json:"nickname"`
	Connects []string `json:"connects"`
}

func NewFriendEvent(nickname []byte) Event {
	return Event{
		Type: NEW_FRIEND_REQ,
		Data: nickname,
	}
}

func AcceptFriendEvent(nickname []byte) Event {
	return Event{
		Type: ACCEPT_FRIEND,
		Data: nickname,
	}
}

func DenyFriendEvent(nickname []byte) Event {
	return Event{
		Type: DENY_FRIEND,
		Data: nickname,
	}
}

func DeleteFriendEvent(nickname []byte) Event {
	return Event{
		Type: DELETE_FRIEND,
		Data: nickname,
	}
}

func BlockUserEvent(nickname []byte) Event {
	return Event{
		Type: BLOCK_USER,
		Data: nickname,
	}
}

func UnblockUserEvent(nickname []byte) Event {
	return Event{
		Type: UNBLOCK_USER,
		Data: nickname,
	}
}

func FriendOnlineEvent(fcd []byte) Event {
	return Event{
		Type: FRIEND_ONLINE,
		Data: fcd,
	}
}

func FriendOfflineEvent(nickname []byte) Event {
	return Event{
		Type: FRIEND_OFFLINE,
		Data: nickname,
	}
}

// func FriendConnectionsEvent(fcd []byte) Event {
// 	return Event{
// 		Type: FRIEND_CONNECTIONS,
// 		Data: fcd,
// 	}
// }
