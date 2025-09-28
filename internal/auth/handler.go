package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte(getJWTSecret())

type User struct {
	ID           string
	Username     string
	PasswordHash string
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "failed to hash password", http.StatusInternalServerError)
			return
		}
		id := uuid.New().String()

		_, err = db.Exec("INSERT INTO users (id, username, password_hash) VALUES (?, ?, ?)",
			id, input.Username, string(hash))
		if err != nil {
			http.Error(w, "username already exists", http.StatusConflict)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "invalid input", http.StatusBadRequest)
			return
		}

		var user User
		err := db.QueryRow("SELECT id, password_hash FROM users WHERE username = ?", input.Username).
			Scan(&user.ID, &user.PasswordHash)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id":  user.ID,
			"username": input.Username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
		tokenString, err := token.SignedString(jwtSecret)
		if err != nil {
			http.Error(w, "failed to sign token", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"token": tokenString,
		})
	}
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "changeme"
	}
	return secret
}
