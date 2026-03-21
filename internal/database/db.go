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

func Open(path string) (*DB, error) {
	dsn := path + "?_foreign_keys=on"
	conn, err := sql.Open("sqlite3", dsn)
		if err != nil{
		fmt.Println("couldn't connect to db:", err)
		return nil, err
	}
	conn.SetMaxOpenConns(1)

	 err = conn.Ping()
	if err != nil{
		fmt.Println("couldn't connect to db:", err)
		return nil, err
	}
	return &DB{db : conn, path: path}, nil
}
