package database

import (
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func (d *DB) CreateUser(u *User) error {
	u.CreatedAt = time.Now()
	query := `INSERT INTO users (name) VALUES(?)`
	res, err :=  d.db.Exec(query, u.Name)
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
