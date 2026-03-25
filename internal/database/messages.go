package database

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID        uuid.UUID `json:"id"         db:"id"`
	SessionID uuid.UUID `json:"session_id" db:"session_id"`
	Role      RoleType  `json:"role"       db:"role"`
	Content   string    `json:"content"    db:"content"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type RoleType string

const (
	UserRole  RoleType = "user"
	ModelRole RoleType = "model"
)
