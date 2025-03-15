package database

//go:generate sqlc generate

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func NewSQLite3DB(dbPath string) (*sql.DB, error) {
	if strings.HasPrefix(dbPath, "sqlite3:") {
		dbPath = strings.TrimPrefix(dbPath, "sqlite3:")
	}
	// Check if the database file exists
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %s", dbPath)
	}

	// Open the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite3 database: %w", err)
	}

	return db, nil
}
