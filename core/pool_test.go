package core

import (
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

func newTestPool() *Pool {
	return &Pool{
		Proxies: []*Proxy{
			newTestProxy("http://127.0.0.1:8888"),
			newTestProxy("http://127.0.0.1:8889"),
		},
	}
}

func newTestProxy(url string) *Proxy {
	return &Proxy{
		URL:       url,
		Alive:     true,
		Score:     6,
		LastTest:  time.Now(),
		Timeout:   2 * time.Second,
		transport: &http.Transport{},
		client:    &http.Client{},
	}
}

func TestAllocate_PrefersHigherScore(t *testing.T) {
	pool := &Pool{
		Proxies: []*Proxy{
			{URL: "http://127.0.0.1:8888", Alive: true, Score: 10},
			{URL: "http://127.0.0.1:8889", Alive: true, Score: 5},
		},
	}

	const trials = 200
	counts := map[string]int{"8888": 0, "8889": 0}

	for range trials {
		p, err := pool.Allocate()
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(p.URL, "8888") {
			counts["8888"]++
		} else {
			counts["8889"]++
		}
	}

	if counts["8888"] < int(0.60*float64(trials)) {
		t.Fatalf("expected proxy 8888 (score 10) to be chosen >=60%% of the time; got: %+v", counts)
	}
}

func TestAllocate_NoAliveProxies(t *testing.T) {
	pool := newTestPool()
	pool.Proxies[0].Alive = false
	pool.Proxies[1].Alive = false

	_, err := pool.Allocate()
	if err == nil {
		t.Error("expected error when all proxies are dead")
	}
}

func TestUsageCountIncrements(t *testing.T) {
	pool := &Pool{
		Proxies: []*Proxy{
			{URL: "http://127.0.0.1:8888", Alive: true, Score: 10},
			{URL: "http://127.0.0.1:8889", Alive: true, Score: 1},
		},
	}

	p1, err := pool.Allocate()
	if err != nil {
		t.Fatal(err)
	}
	if p1.UsageCount != 1 {
		t.Fatalf("expected chosen proxy UsageCount=1 after first allocation, got %d", p1.UsageCount)
	}

	p2, err := pool.Allocate()
	if err != nil {
		t.Fatal(err)
	}
	if !(p2.UsageCount == 1 || p2.UsageCount == 2) {
		t.Fatalf("unexpected UsageCount %d for %s", p2.UsageCount, p2.URL)
	}

	total := pool.Proxies[0].UsageCount + pool.Proxies[1].UsageCount
	if total != 2 {
		t.Fatalf("expected total UsageCount=2, got %d", total)
	}
}

func TestConcurrentAllocation_Safe(t *testing.T) {
	pool := newTestPool()
	var wg sync.WaitGroup
	const goroutines = 100

	wg.Add(goroutines)
	for range goroutines {
		go func() {
			defer wg.Done()
			_, _ = pool.Allocate()
		}()
	}
	wg.Wait()

	totalUsage := pool.Proxies[0].UsageCount + pool.Proxies[1].UsageCount
	if totalUsage != goroutines {
		t.Errorf("expected %d allocations, got %d", goroutines, totalUsage)
	}
}

func TestAllocate_PrefersLowerUsageOnTie(t *testing.T) {
	p1 := &Proxy{URL: "A", Alive: true, Score: 5, UsageCount: 0}
	p2 := &Proxy{URL: "B", Alive: true, Score: 5, UsageCount: 10}

	pool := &Pool{Proxies: []*Proxy{p1, p2}}

	p, err := pool.Allocate()
	if err != nil {
		t.Fatal(err)
	}
	if p.URL != "A" {
		t.Fatalf("expected lower-usage proxy A to be chosen first, got %s", p.URL)
	}

	p, _ = pool.Allocate()
	if p.URL != "A" && p.URL != "B" {
		t.Fatalf("unexpected proxy chosen: %s", p.URL)
	}
}

func TestAllocate_IgnoresDeadProxies(t *testing.T) {
	p1 := &Proxy{URL: "alive", Alive: true, Score: 5}
	p2 := &Proxy{URL: "dead", Alive: false, Score: 100}

	pool := &Pool{Proxies: []*Proxy{p1, p2}}

	p, err := pool.Allocate()
	if err != nil {
		t.Fatal(err)
	}

	if p.URL != "alive" {
		t.Fatalf("expected alive proxy, got %s", p.URL)
	}
}
