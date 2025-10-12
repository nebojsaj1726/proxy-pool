package core

import (
	"sync"
	"testing"
)

func newTestPool() *Pool {
	return &Pool{
		Proxies: []*Proxy{
			{URL: "http://127.0.0.1:8888", Alive: true},
			{URL: "http://127.0.0.1:8889", Alive: true},
		},
	}
}

func TestAllocate_RotatesAliveProxies(t *testing.T) {
	pool := newTestPool()

	first, _ := pool.Allocate()
	second, _ := pool.Allocate()
	third, _ := pool.Allocate()

	if first == nil || second == nil || third == nil {
		t.Fatal("expected non-nil proxies")
	}

	if first.URL == second.URL {
		t.Errorf("expected rotation, got same proxy twice: %s", first.URL)
	}

	if third.URL != first.URL {
		t.Errorf("expected round-robin rotation back to first, got %s", third.URL)
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
	pool := newTestPool()
	proxy, _ := pool.Allocate()
	if proxy.UsageCount != 1 {
		t.Errorf("expected UsageCount=1, got %d", proxy.UsageCount)
	}
	proxy, _ = pool.Allocate()
	if proxy.UsageCount == 0 {
		t.Errorf("expected UsageCount incremented, got %d", proxy.UsageCount)
	}
}

func TestConcurrentAllocation_Safe(t *testing.T) {
	pool := newTestPool()
	wg := sync.WaitGroup{}

	for range 100 {
		wg.Go(func() {
			_, _ = pool.Allocate()
		})
	}
	wg.Wait()

	totalUsage := pool.Proxies[0].UsageCount + pool.Proxies[1].UsageCount
	if totalUsage != 100 {
		t.Errorf("expected 100 allocations, got %d", totalUsage)
	}
}
