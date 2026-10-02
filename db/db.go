package db

import (
	"database/sql"
	_ "embed"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var schema string

var DB *sql.DB

// migrations adds columns that were introduced after the initial schema.
// CREATE TABLE IF NOT EXISTS won't alter an already-existing table, so
// databases created before a given column existed need an explicit
// ALTER TABLE here. Each statement is best-effort: if the column is already
// present, SQLite's "duplicate column name" error is ignored.
var migrations = []string{
	`ALTER TABLE policies ADD COLUMN approved_at TEXT`,
	`ALTER TABLE policies ADD COLUMN approved_by_user_id INTEGER`,
}

func Init(path string) error {
	conn, err := sql.Open("sqlite3", path+"?_foreign_keys=on")
	if err != nil {
		return err
	}
	if _, err := conn.Exec(schema); err != nil {
		return err
	}
	for _, m := range migrations {
		if _, err := conn.Exec(m); err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	DB = conn
	return nil
}
