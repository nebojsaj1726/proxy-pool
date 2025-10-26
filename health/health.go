package health

import (
	"log"
	"time"

	"github.com/nebojsaj1726/proxy-pool/core"
)

type Manager struct {
	Pool     core.Pooler
	Interval time.Duration
	stopCh   chan struct{}
}

func New(pool core.Pooler, interval time.Duration) *Manager {
	return &Manager{
		Pool:     pool,
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
