package olhttp

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestRateLimitedClient_ThrottlesBurst(t *testing.T) {
	// Set up a test server that counts requests.
	var count atomic.Int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Allow 10 rps with burst of 5.
	client := NewClient(10, 5)

	// Fire 5 requests concurrently — they should all succeed within the burst.
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Get(ts.URL)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	if got := count.Load(); got != 5 {
		t.Errorf("expected 5 requests, got %d", got)
	}
}

func TestRateLimitedClient_EnforcesRate(t *testing.T) {
	var count atomic.Int64
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		count.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Very restrictive: 2 rps, burst of 2.
	client := NewClient(2, 2)

	start := time.Now()
	// Send 4 requests sequentially — first 2 should be instant (burst),
	// next 2 should be delayed by ~0.5s each.
	for i := 0; i < 4; i++ {
		resp, err := client.Get(ts.URL)
		if err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
		resp.Body.Close()
	}
	elapsed := time.Since(start)

	if got := count.Load(); got != 4 {
		t.Errorf("expected 4 requests, got %d", got)
	}

	// With burst=2 and rate=2/s, requests 3-4 should take at least ~1s total.
	if elapsed < 800*time.Millisecond {
		t.Errorf("expected rate limiting to impose delay, but elapsed was only %v", elapsed)
	}
}

func TestDefaultClient_ReturnsValidClient(t *testing.T) {
	client := DefaultClient()
	if client == nil {
		t.Fatal("DefaultClient returned nil")
	}
	if client.Timeout != 15*time.Second {
		t.Errorf("expected 15s timeout, got %v", client.Timeout)
	}
}
