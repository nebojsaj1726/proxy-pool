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

type Pooler interface {
	Allocate() (*Proxy, error)
	HealthCheck(timeout time.Duration)
	AliveProxies() []*Proxy
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
