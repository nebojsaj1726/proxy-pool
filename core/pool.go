package core

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"os"

	"gopkg.in/yaml.v3"
)

// Pooler defines the behavior of a proxy pool.
// It allows selecting (allocating) a proxy, running health checks,
// returning alive proxies, and obtaining read-only snapshots.
type Pooler interface {
	Allocate() (*Proxy, error)
	HealthCheck(timeout time.Duration)
	AliveProxies() []*Proxy
	Snapshots() []ProxyStats
}

type Pool struct {
	Proxies []*Proxy
	mu      sync.Mutex
}

type Config struct {
	HealthCheckURL string   `yaml:"health_check_url"`
	TimeoutSeconds int      `yaml:"timeout_seconds"`
	Proxies        []string `yaml:"proxies"`
}

type ProxyStats struct {
	URL          string  `json:"url"`
	Alive        bool    `json:"alive"`
	LastTest     string  `json:"last_test"`
	Score        float64 `json:"score"`
	UsageCount   int     `json:"usage_count"`
	FailCount    int     `json:"fail_count"`
	SuccessCount int     `json:"success_count"`
	LatencyMS    int     `json:"latency_ms"`
}

func LoadConfig(path string) (*Pool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	proxies := make([]*Proxy, 0, len(cfg.Proxies))
	for _, u := range cfg.Proxies {
		proxyURL, err := url.Parse(u)
		if err != nil {
			log.Printf("Skipping invalid proxy URL %s: %v", u, err)
			continue
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		proxies = append(proxies, &Proxy{
			URL:          u,
			Alive:        true,
			LastTest:     time.Now(),
			CheckURL:     cfg.HealthCheckURL,
			Timeout:      time.Duration(cfg.TimeoutSeconds) * time.Second,
			UsageCount:   0,
			FailCount:    0,
			SuccessCount: 0,
			Score:        6,
			transport:    transport,
			client: &http.Client{
				Timeout:   time.Duration(cfg.TimeoutSeconds) * time.Second,
				Transport: transport,
			},
		})
	}

	return &Pool{Proxies: proxies}, nil
}

// Allocate selects the best available proxy.
// Selection rules:
//  1. Only Alive proxies are considered
//  2. Highest Score wins
//  3. On score tie, proxy with lower UsageCount is preferred
//
// Thread-safe.
func (p *Pool) Allocate() (*Proxy, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	type candidate struct {
		proxy *Proxy
		score float64
		use   int
	}

	candidates := make([]candidate, 0, len(p.Proxies))

	for _, proxy := range p.Proxies {
		proxy.mu.Lock()
		if proxy.Alive {
			candidates = append(candidates, candidate{
				proxy: proxy,
				score: proxy.Score,
				use:   proxy.UsageCount,
			})
		}
		proxy.mu.Unlock()
	}

	if len(candidates) == 0 {
		return nil, errors.New("no alive proxies")
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].use < candidates[j].use
		}
		return candidates[i].score > candidates[j].score
	})

	chosen := candidates[0].proxy

	chosen.mu.Lock()
	chosen.UsageCount++
	chosen.mu.Unlock()

	return chosen, nil
}

// HealthCheck performs a concurrent check of all proxies using Proxy.Test().
// It applies score decay, logs status transitions (recovered/degraded),
// and updates latency, score, alive state, etc.
func (p *Pool) HealthCheck(timeout time.Duration) {
	p.mu.Lock()
	proxies := make([]*Proxy, len(p.Proxies))
	copy(proxies, p.Proxies)
	p.mu.Unlock()

	var wg sync.WaitGroup
	for _, proxy := range proxies {
		wg.Add(1)
		go func(pr *Proxy) {
			defer wg.Done()

			pr.DecayScore()
			prevAlive := pr.Snapshot().Alive
			pr.Test(timeout)

			pr.mu.Lock()
			status := pr.Alive
			score := pr.Score
			pr.mu.Unlock()

			if status && !prevAlive {
				log.Printf("Proxy recovered: %s (score=%.1f)", pr.URL, score)
			} else if !status && prevAlive {
				log.Printf("Proxy degraded: %s (score=%.1f)", pr.URL, score)
			} else {
				log.Printf("Proxy check: %s (alive=%t, score=%.1f)", pr.URL, status, score)
			}
		}(proxy)
	}
	wg.Wait()
}

func (p *Pool) AliveProxies() []*Proxy {
	p.mu.Lock()
	defer p.mu.Unlock()
	alive := []*Proxy{}
	for _, proxy := range p.Proxies {
		if proxy.Alive {
			alive = append(alive, proxy)
		}
	}
	return alive
}

func (p *Pool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, proxy := range p.Proxies {
		proxy.Close()
	}
}

func (p *Proxy) RebuildHTTPClient() error {
	proxyURL, err := url.Parse(p.URL)
	if err != nil {
		return err
	}

	transport := &http.Transport{Proxy: http.ProxyURL(proxyURL)}
	p.transport = transport
	p.client = &http.Client{
		Timeout:   p.Timeout,
		Transport: transport,
	}
	return nil
}

func (p *Pool) Snapshots() []ProxyStats {
	p.mu.Lock()
	defer p.mu.Unlock()

	stats := make([]ProxyStats, len(p.Proxies))
	for i, pr := range p.Proxies {
		pr.mu.Lock()
		stats[i] = ProxyStats{
			URL:          pr.URL,
			Alive:        pr.Alive,
			LastTest:     pr.LastTest.Format(time.RFC3339),
			Score:        pr.Score,
			UsageCount:   pr.UsageCount,
			FailCount:    pr.FailCount,
			SuccessCount: pr.SuccessCount,
			LatencyMS:    pr.LatencyMS,
		}
		pr.mu.Unlock()
	}
	return stats
}
