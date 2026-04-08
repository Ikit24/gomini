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
