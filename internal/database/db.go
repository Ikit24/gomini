package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	db *sql.DB
	path string
}

func (d *DB) Close() error {
	return d.db.Close()
}

func Open(path string) (*DB, error) {
	dsn := path + "?_foreign_keys=on"
	conn, err := sql.Open("sqlite3", dsn)
		if err != nil{
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	d := &DB{db: conn, path: path}

	if err := d.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return d, nil
}

func (d *DB) createTables() error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        created_at DATETIME
    );

    CREATE TABLE IF NOT EXISTS messages (
        id TEXT PRIMARY KEY,
        session_id TEXT,
        role TEXT,
        content TEXT,
        created_at DATETIME
    );`

	_, err := d.db.Exec(query)
	return err
}
