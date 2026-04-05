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

func (d *DB) SaveMessage(m *Message) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	query := `INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, m.ID, m.SessionID, m.Role, m.Content, m.CreatedAt)
	return err
}

func (d *DB) GetMessagesBySessionID(sessionID uuid.UUID) ([]Message, error) {
	messages := []Message{}
	query := `SELECT id, session_id, role, content, created_at FROM messages WHERE session_id = ? ORDER BY created_at ASC`

	rows, err := d.db.Query(query, sessionID)
	if err != nil {
		return  nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m Message
		err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

func (d *DB) DeleteSession(sessionID uuid.UUID) error {
	query := `DELETE FROM messages WHERE session_id = ?`

	_, err := d.db.Exec(query, sessionID)
	return err
}
