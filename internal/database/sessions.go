package database

import (
	"fmt"
	"time"
	"errors"
	"database/sql"

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

func (d *DB) UpdateSessionTitle(s *Session) error {
	if s == nil {
		return errors.New("session cannot be nil")
	}
	now := time.Now()
	s.UpdatedAt = now

	query := `UPDATE sessions SET title = ?, updated_at = ? WHERE id = ?`
	res, err := d.db.Exec(query, s.Title, s.UpdatedAt, s.ID)
	if err != nil {
		return fmt.Errorf("couldn't update session title: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return errors.New("no session found with that ID")
	}

	return nil
}

func (d *DB) GetSessionByID(ID uuid.UUID) (*Session, error) {
	var s Session
	query := `SELECT id, user_id, title, created_at, updated_at FROM sessions WHERE id = ?`

	err := d.db.QueryRow(query, ID).Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows{
			return nil, fmt.Errorf("lookup failed: %w", sql.ErrNoRows)
		}
		return nil, err
	}

	return &s, nil
}
