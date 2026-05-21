package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID       uuid.UUID
	MessagesChan chan SessionMessage
	CreatedTime  time.Time
}

const (
	NEW_FRIEND = iota
	ACCEPT_FRIEND
	DENY_FRIEND
)

type SessionMessage struct {
	Type uint
	Data []byte
}
