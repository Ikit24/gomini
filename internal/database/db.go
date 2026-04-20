package database

import (
	"database/sql"

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

	return &DB{db: conn}, nil
}
