package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Nickname     string
	Key          string
	Fingerprint  string
	RegisterTime time.Time
}