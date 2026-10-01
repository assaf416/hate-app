package db

import (
	"database/sql"
	_ "embed"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

var DB *sql.DB

func Init(path string) error {
	conn, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(schema); err != nil {
		return err
	}
	DB = conn
	return nil
}
