package db

import (
	"database/sql"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
)

func ConnectAndMigrate() *Store {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./proxy-pool.db"
	}
	firstRun := false

	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		firstRun = true
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("failed to open database:", err)
	}

	if firstRun {
		log.Println("No database found, running migrations...")
		m, err := migrate.New(
			"file://migrations",
			"sqlite3://"+dbPath,
		)
		if err != nil {
			log.Fatal("failed to load migrations:", err)
		}
		if err := m.Up(); err != nil && err.Error() != "no change" {
			log.Fatal("failed to apply migrations:", err)
		}
		log.Println("Migrations applied.")
	} else {
		log.Println("Database already exists, skipping migrations.")
	}

	return &Store{DB: db}
}

func (s *Store) CreateUser(id, username, passwordHash string) error {
	_, err := s.DB.Exec(
		"INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)",
		id, username, passwordHash,
	)
	return err
}

func (s *Store) GetUserByUsername(username string) (id string, passwordHash string, err error) {
	err = s.DB.QueryRow(
		"SELECT id, password_hash FROM users WHERE username = ?",
		username,
	).Scan(&id, &passwordHash)
	return
}
