package database

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Message struct {
	ID uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}
