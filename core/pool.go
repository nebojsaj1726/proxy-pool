package core

import (
	"errors"
	"log"
	"net/http"
	"net/url"
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

	proxies := make([]*Proxy, len(cfg.Proxies))
	for i, u := range cfg.Proxies {
		proxyURL, err := url.Parse(u)
		if err != nil {
			log.Printf("Skipping invalid proxy URL %s: %v", u, err)
			continue
		}

		transport := &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		}

		proxies[i] = &Proxy{
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
		}
	}

	return &Pool{Proxies: proxies}, nil
}

func (p *Pool) Allocate() (*Proxy, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var chosen *Proxy
	highestScore := -1000
	minUsage := int(^uint(0) >> 1)

	for _, proxy := range p.Proxies {
		proxy.mu.Lock()
		score := proxy.Score
		alive := proxy.Alive
		usage := proxy.UsageCount
		proxy.mu.Unlock()

		if !alive || score <= 0 {
			continue
		}

		if score > highestScore || (score == highestScore && usage < minUsage) {
			chosen = proxy
			highestScore = score
			minUsage = usage
		}
	}

	if chosen == nil {
		return nil, errors.New("no alive proxies")
	}

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
			prevAlive := pr.Alive
			pr.Test(timeout)

			pr.mu.Lock()
			status := pr.Alive
			score := pr.Score
			pr.mu.Unlock()

			if status && !prevAlive {
				log.Printf("Proxy recovered: %s (score=%d)", pr.URL, score)
			} else if !status && prevAlive {
				log.Printf("Proxy degraded: %s (score=%d)", pr.URL, score)
			} else {
				log.Printf("Proxy check: %s (alive=%t, score=%d)", pr.URL, status, score)
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
