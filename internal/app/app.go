package app

import (
	"context"
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

type App struct {
	DB     db.UserStore
	Pool   core.Pooler
	Mux    *http.ServeMux
	Server *http.Server
}

func NewApp() (*App, error) {
	_ = godotenv.Load()

	database := db.ConnectAndMigrate()

	pool, err := core.LoadConfig("./config.yaml")
	if err != nil {
		return nil, err
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
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	protected := http.NewServeMux()
	protected.Handle("/proxies", api.ListProxiesHandler(pool))
	protected.Handle("/allocate", api.AllocateProxyHandler(pool))

	mux.Handle("/proxies", auth.JWTMiddleware(protected))
	mux.Handle("/allocate", auth.JWTMiddleware(protected))

	server := &http.Server{
		Addr:    ":8080",
		Handler: middleware.LoggingMiddleware(mux),
	}

	return &App{
		DB:     database,
		Pool:   pool,
		Mux:    mux,
		Server: server,
	}, nil
}

func (a *App) Start() error {
	log.Println("Server running on :8080")
	return a.Server.ListenAndServe()
}

func (a *App) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = a.Server.Shutdown(ctx)
	log.Println("Server stopped gracefully")
}
