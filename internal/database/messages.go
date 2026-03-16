package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	SessionID uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Role RoleType       `json:"role" db:"role_column"`
	Content string      `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type RoleType string

const (
	UserRole  RoleType = "user"
	ModelRole RoleType = "model"
)
