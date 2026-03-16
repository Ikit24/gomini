package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID uuid.UUID        `json:"id"`
	Title string        `json:"title" db:"chat_title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
