package database

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID `json:"id" db:"id"`
	UserID    uuid.UUID `json:"user_id" db:"user_id"`
	Title     string    `json:"title" db:"title"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

func (d *DB) CreateSession(s *Session) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	now := time.Now()
	s.ID = id
	s.CreatedAt = now
	s.UpdatedAt = now

	query := `INSERT INTO sessions (id, user_id, title, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`
	_, err = d.db.Exec(query, s.ID, s.UserID, s.Title, s.CreatedAt, s.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) DeleteSession(sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = ?`

	_, err := d.db.Exec(query, sessionID)
	return err
}

func (d *DB) GetSessionsByUserID(userID uuid.UUID) ([]Session, error) {
	sessions := []Session{}
	query := `SELECT id, user_id, title, created_at, updated_at FROM sessions WHERE user_id = ? ORDER BY created_at DESC`

	rows, err := d.db.Query(query, userID)
	if err != nil {
		return  nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var s Session
		err := rows.Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (d *DB) CreateMessage(m *Message) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	now := time.Now()
	m.ID = id
	m.CreatedAt = now

	query := `INSERT INTO messages (id, session_id, role, content, created_at) VALUES (?, ?, ?, ?, ?)`
	_, err = d.db.Exec(query, m.ID, m.SessionID, m.Role, m.Content, m.CreatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) UpdateSessionTitle(s *Session) error {
	now := time.Now()
	s.UpdatedAt = now

	query := `UPDATE sessions SET title = ?, updated_at = ? WHERE id = ?`
	_, err := d.db.Exec(query, s.Title, s.UpdatedAt, s.ID)
	if err != nil {
		return err
	}

	return nil
}
