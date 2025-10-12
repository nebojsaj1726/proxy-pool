package core

import (
	"errors"
	"log"
	"sync"
	"time"

	"os"

	"gopkg.in/yaml.v3"
)

type Pool struct {
	Proxies []*Proxy
	mu      sync.Mutex
	index   int
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
		proxies[i] = &Proxy{
			URL:        u,
			Alive:      true,
			LastTest:   time.Now(),
			CheckURL:   cfg.HealthCheckURL,
			Timeout:    time.Duration(cfg.TimeoutSeconds) * time.Second,
			UsageCount: 0,
			FailCount:  0,
		}
	}

	return &Pool{Proxies: proxies, index: 0}, nil
}

func (p *Pool) Allocate() (*Proxy, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var chosen *Proxy
	minUsage := int(^uint(0) >> 1)

	for _, proxy := range p.Proxies {
		if proxy.Alive && proxy.UsageCount <= minUsage {
			if proxy.UsageCount < minUsage || chosen == nil {
				chosen = proxy
				minUsage = proxy.UsageCount
			}
		}
	}

	if chosen == nil {
		return nil, errors.New("no alive proxies")
	}

	chosen.UsageCount++
	p.index = (p.index + 1) % len(p.Proxies)
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
			if pr.Test(timeout) {
				pr.FailCount = 0
				pr.Alive = true
				log.Printf("Health check OK: %s", pr.URL)
			} else {
				pr.FailCount++
				if pr.FailCount >= 3 {
					pr.Alive = false
					log.Printf("Proxy marked dead after 3 fails: %s", pr.URL)
				}
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
