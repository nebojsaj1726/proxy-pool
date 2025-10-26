package core

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Proxy struct {
	URL        string
	Alive      bool
	LastTest   time.Time
	CheckURL   string
	Timeout    time.Duration
	UsageCount int
	FailCount  int
	mu         sync.Mutex
	transport  *http.Transport
	client     *http.Client
}

type ProxySnapshot struct {
	URL      string
	Alive    bool
	LastTest time.Time
}

func (p *Proxy) Test(timeout time.Duration) bool {
	p.mu.Lock()
	client := p.client
	checkURL := p.CheckURL
	p.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		p.mu.Lock()
		p.Alive = false
		p.LastTest = time.Now()
		p.mu.Unlock()
		log.Printf("Health check request creation failed for %s: %v", p.URL, err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		p.mu.Lock()
		p.Alive = false
		p.LastTest = time.Now()
		p.mu.Unlock()
		log.Printf("Health check failed for %s: %v", p.URL, err)

		return false
	}

	defer resp.Body.Close()
	alive := resp.StatusCode >= 200 && resp.StatusCode < 300

	p.mu.Lock()
	p.Alive = alive
	p.LastTest = time.Now()
	p.mu.Unlock()
	return alive
}

func (p *Proxy) Snapshot() ProxySnapshot {
	p.mu.Lock()
	defer p.mu.Unlock()
	return ProxySnapshot{
		URL:      p.URL,
		Alive:    p.Alive,
		LastTest: p.LastTest,
	}
}

func (p *Proxy) Close() {
	if p.transport != nil {
		p.transport.CloseIdleConnections()
	}
}
