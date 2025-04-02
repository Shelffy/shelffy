package auth

import (
	"github.com/google/uuid"
	"time"
)

type Session struct {
	UserID    uuid.UUID
	IsActive  bool
	ExpiresAt time.Time
	ID        string
}
