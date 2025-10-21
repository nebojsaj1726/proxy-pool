package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

type mockStore struct {
	users map[string]string
}

func newMockStore() *mockStore {
	return &mockStore{users: make(map[string]string)}
}

func (m *mockStore) CreateUser(id, username, passwordHash string) error {
	if _, exists := m.users[username]; exists {
		return fmt.Errorf("username exists")
	}
	m.users[username] = passwordHash
	return nil
}

func (m *mockStore) GetUserByUsername(username string) (string, string, error) {
	hash, ok := m.users[username]
	if !ok {
		return "", "", fmt.Errorf("user not found")
	}
	return "mock-id", hash, nil
}

func TestRegisterAndLogin(t *testing.T) {
	os.Setenv("JWT_SECRET", "testsecret")
	store := newMockStore()

	regBody := `{"username":"alice","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString(regBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	RegisterHandler(store).ServeHTTP(w, req)

	if w.Result().StatusCode != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Result().StatusCode)
	}

	loginBody := `{"username":"alice","password":"secret"}`
	req = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(loginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	LoginHandler(store).ServeHTTP(w, req)

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
