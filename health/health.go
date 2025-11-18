package health

import (
	"log"
	"time"

	"github.com/nebojsaj1726/proxy-pool/core"
	"github.com/nebojsaj1726/proxy-pool/db"
)

type Manager struct {
	Pool     core.Pooler
	Store    *db.Store
	Interval time.Duration
	stopCh   chan struct{}
}

func New(pool core.Pooler, store *db.Store, interval time.Duration) *Manager {
	return &Manager{
		Pool:     pool,
		Store:    store,
		Interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (m *Manager) Start() {
	log.Printf("[health] starting background checks every %s", m.Interval)
	ticker := time.NewTicker(m.Interval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				start := time.Now()
				m.Pool.HealthCheck(3 * time.Second)

				if m.Store != nil {
					if pool, ok := m.Pool.(*core.Pool); ok {
						for _, pr := range pool.Proxies {
							if err := m.Store.SaveProxy(pr); err != nil {
								log.Printf("[warn] failed to persist proxy %s: %v", pr.URL, err)
							}
						}
					}
				}

				alive := len(m.Pool.AliveProxies())
				total := alive + countDead(m.Pool)
				duration := time.Since(start)

				log.Printf("[health] check complete — alive: %d / %d, duration: %s", alive, total, duration)

			case <-m.stopCh:
				log.Println("[health] stopping background checks")
				return
			}
		}
	}()
}

func (m *Manager) Stop() {
	close(m.stopCh)
}

func countDead(pool core.Pooler) int {
	all := pool.AliveProxies()
	if p, ok := pool.(*core.Pool); ok {
		return len(p.Proxies) - len(all)
	}
	return 0
}
