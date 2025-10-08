package entities

import (
	"time"

	"github.com/google/uuid"
)

type BookHash [256 / 8]byte

type Book struct {
	ID          uuid.UUID
	Title       string
	StoragePath string
	Hash        BookHash
	UploadedBy  uuid.UUID
	UploadedAt  time.Time
}
