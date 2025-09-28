package auth

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	dbPath := "file::memory:?cache=shared"
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open in-memory db: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	if err != nil {
		t.Fatalf("failed to migrate schema: %v", err)
	}
	return db
}

func TestRegisterAndLogin(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	db := setupTestDB(t)
	defer db.Close()

	regBody := `{"username":"alice","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(db).ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Result().StatusCode)
	}

	loginBody := `{"username":"alice","password":"secret"}`
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	LoginHandler(db).ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Result().StatusCode)
	}

	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if resp["token"] == "" {
		t.Fatal("expected JWT token in response")
	}
}
