package core

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"
)

// Proxy represents a single upstream proxy and maintains:
//   - health information
//   - score (quality/priority)
//   - latency statistics
//   - usage counts
//   - internal http.Client configured to route through the proxy
//
// All mutable fields are protected by the internal mutex (mu).
type Proxy struct {
	URL          string
	Alive        bool
	LastTest     time.Time
	CheckURL     string
	Timeout      time.Duration
	UsageCount   int
	FailCount    int
	SuccessCount int
	Score        float64
	mu           sync.Mutex
	transport    *http.Transport
	client       *http.Client
	LatencyMS    int
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

	if client == nil {
		p.recordFailure("no http client", nil)
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, checkURL, nil)
	if err != nil {
		p.recordFailure("request creation failed", err)
		return false
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start).Milliseconds()

	if err != nil {
		p.recordFailure("request failed", err)
		return false
	}
	defer resp.Body.Close()

	ok := resp.StatusCode >= 200 && resp.StatusCode < 300
	if ok {
		p.recordSuccess(int(latency))
	} else {
		p.recordFailure("non-200 status", nil)
	}
	return ok
}

func (p *Proxy) recordSuccess(latencyMS int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	const successGain = 0.4
	const decay = 0.995
	const maxScore = 10.0

	p.SuccessCount++
	p.LatencyMS = latencyMS

	p.Score = p.Score*decay + successGain
	if p.Score > maxScore {
		p.Score = maxScore
	}

	p.Alive = true
	p.LastTest = time.Now()
}

func (p *Proxy) recordFailure(reason string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	const failurePenalty = 0.7
	const decay = 0.995
	const minScore = -5.0
	const softCap = 3

	p.FailCount++

	penalty := failurePenalty
	if p.FailCount <= softCap {
		penalty *= 0.5
	}

	p.Score = p.Score*decay - penalty
	if p.Score < minScore {
		p.Score = minScore
	}

	p.Alive = false
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
	const decayStep = 0.5

	elapsed := time.Since(p.LastTest)
	if elapsed < decayInterval {
		return
	}

	decayPoints := float64(elapsed / decayInterval)
	p.Score -= decayPoints * decayStep

	if p.Score < -5 {
		p.Score = -5
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
