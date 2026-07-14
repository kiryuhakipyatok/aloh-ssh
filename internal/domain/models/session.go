package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID          uuid.UUID
	EventsChan      chan Event
	CurrentConnects []Identity
	CreatedTime     time.Time
}