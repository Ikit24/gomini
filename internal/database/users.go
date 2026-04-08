package database

import (
	"time"
	"database/sql"


	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

func (d *DB) CreateUser(u *User) error {
	id, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	now := time.Now()
	u.ID = id
	u.CreatedAt = now
	u.UpdatedAt = now

	query := `INSERT INTO users (id, name, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err =  d.db.Exec(query, u.ID, u.Name, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

func (d *DB) GetUserByName(name string) (*User, error) {
	var u User
	query := `SELECT id, name, created_at, updated_at FROM users WHERE name = ?`

	row := d.db.QueryRow(query, name)

	err := row.Scan(&u.ID, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
