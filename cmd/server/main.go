package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/nebojsaj1726/proxy-pool/internal/api"
	"github.com/nebojsaj1726/proxy-pool/internal/auth"
	"github.com/nebojsaj1726/proxy-pool/internal/core"
	"github.com/nebojsaj1726/proxy-pool/internal/db"
	"github.com/nebojsaj1726/proxy-pool/internal/middleware"
)

type Proxy struct {
	ID     string `json:"id"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Proto  string `json:"proto"`
	Status string `json:"status"`
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system env")
	}

	database := db.ConnectAndMigrate()

	pool, err := core.LoadConfig("./config.yaml")
	if err != nil {
		log.Fatal("failed to load proxy config:", err)
	}

	go func() {
		for {
			pool.HealthCheck(3 * time.Second)
			time.Sleep(10 * time.Second)
		}
	}()

	mux := http.NewServeMux()

	mux.Handle("/auth/register", auth.RegisterHandler(database))
	mux.Handle("/auth/login", auth.LoginHandler(database))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	protected := http.NewServeMux()
	protected.Handle("/proxies", api.ListProxiesHandler(pool))
	protected.Handle("/allocate", api.AllocateProxyHandler(pool))

	mux.Handle("/proxies", auth.JWTMiddleware(protected))
	mux.Handle("/allocate", auth.JWTMiddleware(protected))

	loggedMux := middleware.LoggingMiddleware(mux)

	log.Println("Server running on :8080")
	if err := http.ListenAndServe(":8080", loggedMux); err != nil {
		log.Fatal("Server error:", err)
	}
}
