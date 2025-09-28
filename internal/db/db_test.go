package db

import (
	"database/sql"
	"os"
	"testing"
)

func TestInMemoryDB(t *testing.T) {
	os.Setenv("DB_PATH", "file::memory:?cache=shared")

	db, err := sql.Open("sqlite3", os.Getenv("DB_PATH"))
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to create users table: %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}
}
