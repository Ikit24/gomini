package database

import (
	"database/sql"
	"embed"

	_ "modernc.org/sqlite"
	"github.com/pressly/goose/v3"
)

type DB struct {
	db   *sql.DB
	path string
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) Ping() error {
	return d.db.Ping()
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func Open(path string) (*DB, error) {
	dsn := path + "?_pragma=foreign_keys(on)"

	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	var fkStatus int
	err = conn.QueryRow("PRAGMA foreign_keys").Scan(&fkStatus)
	// fkStatus should be 1 if it's working
	conn.SetMaxOpenConns(1)

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, err
	}
	
	if err := goose.Up(conn, "migrations"); err != nil {
		return nil, err
	}

	return &DB{db: conn}, nil
}
