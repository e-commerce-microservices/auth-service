// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.15.0

package repository

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID           uuid.UUID
	UserID       int64
	RefreshToken string
	UserAgent    string
	ClientIp     string
	ExpiresAt    time.Time
	CreatedAt    time.Time
}
