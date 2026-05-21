package models

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	UserID      uuid.UUID
	CreatedTime time.Time
}
