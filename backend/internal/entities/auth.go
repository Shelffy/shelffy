package entities

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	IsActive  bool
	ExpiresAt time.Time
}

var NilSession = Session{}
