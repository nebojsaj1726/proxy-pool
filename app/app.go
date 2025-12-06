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
	DB     *db.Store
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

	storedProxies, err := database.LoadProxies()
	if err != nil {
		log.Printf("warning: failed to load proxies from DB: %v", err)
	} else {
		for _, p := range storedProxies {
			if len(pool.Proxies) > 0 {
				p.CheckURL = pool.Proxies[0].CheckURL
				p.Timeout = pool.Proxies[0].Timeout
			}

			if err := p.RebuildHTTPClient(); err != nil {
				log.Printf("Skipping invalid DB proxy URL %s: %v", p.URL, err)
				continue
			}
		}

		merged := make([]*core.Proxy, 0)
		seen := make(map[string]bool)

		for _, p := range storedProxies {
			merged = append(merged, p)
			seen[p.URL] = true
		}

		for _, p := range pool.Proxies {
			if !seen[p.URL] {
				merged = append(merged, p)
			}
		}

		pool.Proxies = merged
	}

	healthManager := health.New(pool, database, 5*time.Second)
	healthManager.Start()

	mux := http.NewServeMux()

	mux.Handle("/auth/register", auth.RegisterHandler(database))
	mux.Handle("/auth/login", auth.LoginHandler(database))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"ok"}`)); err != nil {
			http.Error(w, "failed to write response", http.StatusInternalServerError)
			return
		}
	})

	protected := http.NewServeMux()
	protected.Handle("/proxies", api.ListProxiesHandler(pool))
	protected.Handle("/proxies/stats", api.StatsHandler(pool))
	protected.Handle("/allocate", api.AllocateProxyHandler(pool))

	mux.Handle("/proxies", auth.JWTMiddleware(protected))
	mux.Handle("/proxies/stats", auth.JWTMiddleware(protected))
	mux.Handle("/allocate", auth.JWTMiddleware(protected))

	server := &http.Server{
		Addr: ":8080",
		Handler: middleware.CORS(
			middleware.LoggingMiddleware(mux)),
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

	if a.Pool != nil {
		if pool, ok := a.Pool.(*core.Pool); ok {
			pool.Close()
			log.Println("[pool] all proxy connections closed")

			if a.DB != nil {
				a.DB.SaveAllProxies(pool.Proxies)
				log.Println("[db] all proxies persisted")
			}
		}
	}

	log.Println("Server stopped gracefully")
}
