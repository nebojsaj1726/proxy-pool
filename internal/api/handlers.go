package api

import (
	"encoding/json"
	"net/http"

	"github.com/nebojsaj1726/proxy-pool/internal/core"
)

func ListProxiesHandler(pool *core.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type proxyInfo struct {
			URL      string `json:"url"`
			Alive    bool   `json:"alive"`
			LastTest string `json:"last_test"`
		}

		alive := pool.AliveProxies()
		resp := make([]proxyInfo, len(alive))
		for i, p := range alive {
			resp[i] = proxyInfo{
				URL:      p.URL,
				Alive:    p.Alive,
				LastTest: p.LastTest.Format("15:04:05"),
			}
		}

		json.NewEncoder(w).Encode(resp)
	}
}

func AllocateProxyHandler(pool *core.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxy, err := pool.Allocate()
		if err != nil {
			http.Error(w, "no alive proxies", http.StatusServiceUnavailable)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{
			"allocated": proxy.URL,
		})
	}
}
