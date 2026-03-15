package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uuid.UUID `json:"id"`
	SessionID uuid.UUID `json:"session_id"`
	Role RoleType `json:"role" db:"role_column"`
	Content string
	CreatedAt time.Time `json:"created_at"`
}
