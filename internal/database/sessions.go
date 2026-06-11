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

var ErrNotFound  = errors.New("resource not found or unauthorized")

func (d *DB) CreateSession(s *Session) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	now := time.Now().UTC()
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

func (d *DB) GetSessionByID(id uuid.UUID) (*Session, error) {
	var s Session
	query := `SELECT id, user_id, title, created_at, updated_at FROM sessions WHERE id = ?`

	err := d.db.QueryRow(query, id).Scan(&s.ID, &s.UserID, &s.Title, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows{
			return nil, fmt.Errorf("lookup failed: %w", sql.ErrNoRows)
		}
		return nil, err
	}

	return &s, nil
}

func (d *DB) UpdateSession(id uuid.UUID, title string) error {
	query := `UPDATE sessions SET title = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := d.db.Exec(query, title, id)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) GetAllSessions() ([]Session, error) {
	var s []Session

	query := `SELECT id, user_id, title, created_at, updated_at FROM sessions`

	rows, err := d.db.Query(query)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

	for rows.Next() {
		var currentSession Session
		err = rows.Scan(&currentSession.ID, &currentSession.UserID, &currentSession.Title, &currentSession.CreatedAt, &currentSession.UpdatedAt)
		if err != nil {
			return nil, err
			}
		s = append(s, currentSession)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return s, nil
}

func (d *DB) DeleteSessionBySessionID(sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE id = ?`

	res, err := d.db.Exec(query, sessionID)
	if  err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

func (d *DB) SaveSession(m *Session) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	if m.CreatedAt.IsZero() {
		m.CreatedAt = time.Now()
	}

	query := `INSERT INTO sessions (id, user_id, title, created_at, updated_at) VALUES (?, ?, ?, ?, ?)`

	_, err := d.db.Exec(query, m.ID, m.UserID, m.Title, m.CreatedAt, m.UpdatedAt)
	return err
}
