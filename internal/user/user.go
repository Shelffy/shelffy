package user

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	Email     string    `json:"email,omitempty"`
	Password  string    `json:"password,omitempty"`
	IsActive  bool      `json:"is_active,omitempty"`
}
