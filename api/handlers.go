package api

import (
	"encoding/json"
	"net/http"

	"github.com/nebojsaj1726/proxy-pool/core"
)

func ListProxiesHandler(pool core.Pooler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type proxyInfo struct {
			URL      string `json:"url"`
			Alive    bool   `json:"alive"`
			LastTest string `json:"last_test"`
		}

		w.Header().Set("Content-Type", "application/json")
		alive := pool.AliveProxies()
		resp := make([]proxyInfo, len(alive))
		for i, p := range alive {
			snap := p.Snapshot()
			resp[i] = proxyInfo{
				URL:      snap.URL,
				Alive:    snap.Alive,
				LastTest: snap.LastTest.Format("15:04:05"),
			}
		}

		_ = json.NewEncoder(w).Encode(resp)
	}
}

func AllocateProxyHandler(pool core.Pooler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		proxy, err := pool.Allocate()
		if err != nil {
			http.Error(w, "no alive proxies", http.StatusServiceUnavailable)
			return
		}

		_ = json.NewEncoder(w).Encode(map[string]string{
			"allocated": proxy.URL,
		})
	}
}

func StatsHandler(pool core.Pooler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats := pool.Snapshots()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(stats)
	})
}
