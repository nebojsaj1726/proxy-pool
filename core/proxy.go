package core

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

type Proxy struct {
	URL          string
	Alive        bool
	LastTest     time.Time
	CheckURL     string
	Timeout      time.Duration
	UsageCount   int
	FailCount    int
	SuccessCount int
	Score        int
	mu           sync.Mutex
	transport    *http.Transport
	client       *http.Client
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
		p.recordFailure("request creation failed", err)
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		p.recordFailure("request failed", err)
		return false
	}

	defer resp.Body.Close()
	alive := resp.StatusCode >= 200 && resp.StatusCode < 300

	if alive {
		p.recordSuccess()
	} else {
		p.recordFailure("non-200 status", nil)
	}
	return alive
}

func (p *Proxy) recordSuccess() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.SuccessCount++
	p.Score++
	if p.Score > 10 {
		p.Score = 10
	}
	p.Alive = p.Score > 0
	p.LastTest = time.Now()
}

func (p *Proxy) recordFailure(reason string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.FailCount++
	p.Score -= 2
	if p.Score < -5 {
		p.Score = -5
	}
	p.Alive = p.Score > 0
	p.LastTest = time.Now()

	if err != nil {
		log.Printf("Proxy check failed for %s: %s (%v)", p.URL, reason, err)
	} else {
		log.Printf("Proxy check failed for %s: %s", p.URL, reason)
	}
}

func (p *Proxy) DecayScore() {
	p.mu.Lock()
	defer p.mu.Unlock()

	const decayInterval = 10 * time.Minute

	inactive := time.Since(p.LastTest)
	decayPoints := int(inactive / decayInterval)
	if decayPoints > 0 {
		p.Score -= decayPoints
		if p.Score < -5 {
			p.Score = -5
		}
		p.Alive = p.Score > 0
		p.LastTest = time.Now()
	}
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
