package auth

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	ID        string
	UserID    uuid.UUID
	IsActive  bool
	ExpiresAt time.Time
}

var NilSession = Session{}
