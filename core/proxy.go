package core

import (
	"log"
	"net/http"
	"net/url"
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
}

type ProxySnapshot struct {
	URL      string
	Alive    bool
	LastTest time.Time
}

func (p *Proxy) Test(timeout time.Duration) bool {
	proxyURL, err := url.Parse(p.URL)
	if err != nil {
		p.Alive = false
		p.LastTest = time.Now()
		return false
	}

	client := &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(p.CheckURL)
	if err != nil {
		log.Printf("Health check failed for %s: %v", p.URL, err)
		p.mu.Lock()
		p.Alive = false
		p.LastTest = time.Now()
		p.mu.Unlock()
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
