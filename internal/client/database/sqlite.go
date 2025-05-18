package database

//go:generate sqlc generate

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema.sql
var embeddedSchema string

var (
	ErrDBNotFound    = fmt.Errorf("database not found")
	ErrAbortedByUser = fmt.Errorf("aborted by user")
)

func NewSQLite3DB(dbPath string) (*sql.DB, error) {
	// Check if the database file exists
	_, err := os.Stat(dbPath)
	if os.IsNotExist(err) {
		return nil, ErrDBNotFound
	}

	// Open the SQLite3 database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite3 database: %w", err)
	}

	return db, nil
}

func CreateSQLite3DB(dbPath string) (*sql.DB, error) {
	// Create the SQLite3 database file
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create SQLite3 database: %w", err)
	}

	// Execute the embedded schema
	stmts := strings.Split(embeddedSchema, ";")
	for _, stmt := range stmts {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		_, err := db.Exec(stmt)
		if err != nil {
			db.Close()
			return nil, fmt.Errorf("failed to execute schema statement: %w\nSQL: %s", err, stmt)
		}
	}

	return db, nil
}

func GetOrCreateSQLite3DB(dbPath string) (*sql.DB, error) {
	db, err := NewSQLite3DB(dbPath)
	if err != nil {
		if errors.Is(err, ErrDBNotFound) {
			fmt.Printf(
				"Database not found at %q. Would you like to create it? (y/N): ",
				dbPath,
			)
			var response string
			_, scanErr := fmt.Scanln(&response)
			if scanErr != nil || (response != "y" && response != "Y") {
				return nil, ErrAbortedByUser
			}
			db, err = CreateSQLite3DB(dbPath)
			if err != nil {
				return nil, fmt.Errorf("CreateSQLite3DB: %w", err)
			}
			return db, nil
		}
		return nil, fmt.Errorf("NewSQLite3DB: %w", err)
	}
	return db, nil
}
