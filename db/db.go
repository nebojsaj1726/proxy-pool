package db

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/mattn/go-sqlite3"
	"github.com/nebojsaj1726/proxy-pool/core"
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

func (s *Store) SaveProxy(p *core.Proxy) error {
	_, err := s.DB.Exec(`
		INSERT INTO proxies (url, score, alive, last_test, usage_count, fail_count, success_count, latency_ms)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			score = excluded.score,
			alive = excluded.alive,
			last_test = excluded.last_test,
			usage_count = excluded.usage_count,
			fail_count = excluded.fail_count,
			success_count = excluded.success_count,
			latency_ms = excluded.latency_ms
	`, p.URL, p.Score, p.Alive, p.LastTest, p.UsageCount, p.FailCount, p.SuccessCount, p.LatencyMS)
	return err
}

func (s *Store) SaveAllProxies(proxies []*core.Proxy) {
	for _, p := range proxies {
		if err := s.SaveProxy(p); err != nil {
			println("[warn] failed to save proxy:", p.URL, "err:", err.Error())
		}
	}
}

func (s *Store) LoadProxies() ([]*core.Proxy, error) {
	rows, err := s.DB.Query(`
		SELECT url, score, alive, last_test, usage_count, fail_count, success_count, latency_ms
		FROM proxies
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var proxies []*core.Proxy
	for rows.Next() {
		var p core.Proxy
		var lastTest time.Time

		if err := rows.Scan(
			&p.URL, &p.Score, &p.Alive, &lastTest,
			&p.UsageCount, &p.FailCount, &p.SuccessCount, &p.LatencyMS,
		); err != nil {
			return nil, err
		}

		p.LastTest = lastTest
		p.Timeout = 5 * time.Second

		proxies = append(proxies, &p)
	}

	return proxies, nil
}
