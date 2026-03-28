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

func SaveMessage(m *Message) error {
	if m.ID == uuid.Nil {
		m.uuid.New()
	}

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	query := `INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, m.ID, m.SessionID, m.Role, m.Content, m.CreatedAt)
	return err
}
