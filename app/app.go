package app

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/nebojsaj1726/proxy-pool/api"
	"github.com/nebojsaj1726/proxy-pool/auth"
	"github.com/nebojsaj1726/proxy-pool/core"
	"github.com/nebojsaj1726/proxy-pool/db"
	"github.com/nebojsaj1726/proxy-pool/health"
	"github.com/nebojsaj1726/proxy-pool/middleware"
)

type App struct {
	DB     db.UserStore
	Pool   core.Pooler
	Mux    *http.ServeMux
	Server *http.Server
	Health *health.Manager
}

func NewApp() (*App, error) {
	_ = godotenv.Load()

	database := db.ConnectAndMigrate()

	pool, err := core.LoadConfig("./config.yaml")
	if err != nil {
		return nil, err
	}

	healthManager := health.New(pool, 5*time.Second)
	healthManager.Start()

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
		Health: healthManager,
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

	if a.Health != nil {
		a.Health.Stop()
	}

	log.Println("Server stopped gracefully")
}
