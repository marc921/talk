package database

//go:generate sqlc generate

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLite3DB(dbPath string) (*sql.DB, error) {
	// Open the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite3 database: %w", err)
	}

	// Check if the database file exists
	_, err = os.Stat(dbPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", dbPath)
	}

	return db, nil
}
