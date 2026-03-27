package database

import (
	"time"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func (d *DB) CreateUser(u *User) error {
	u.CreatedAt = time.Now()
	query := `INSERT INTO users (name, created_at) VALUES (?, ?)`
	res, err :=  d.db.Exec(query, u.Name, u.CreatedAt)
	if err != nil {
		return err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	u.ID = int(id)

	return nil
}

func (d *DB) GetUserByName(name string) (*User, error) {
	var u User
	query := `SELECT id, name, created_at FROM users WHERE name = ?`

	row := d.db.QueryRow(query, name)

	err := row.Scan(&u.ID, &u.Name, &u.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &u, nil
}
