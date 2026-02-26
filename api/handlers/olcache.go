package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

const olCacheTTL = 24 * time.Hour

// cacheEntry holds a cached response with its expiration time.
type cacheEntry struct {
	data      []byte
	expiresAt time.Time
}

// olCache is a simple in-memory TTL cache for Open Library API responses.
type olCache struct {
	entries sync.Map
	hits    atomic.Int64
	misses  atomic.Int64
}

func (c *olCache) get(key string) ([]byte, bool) {
	val, ok := c.entries.Load(key)
	if !ok {
		c.misses.Add(1)
		return nil, false
	}
	entry := val.(cacheEntry)
	if time.Now().After(entry.expiresAt) {
		c.entries.Delete(key)
		c.misses.Add(1)
		return nil, false
	}
	c.hits.Add(1)
	return entry.data, true
}

func (c *olCache) set(key string, data []byte) {
	c.entries.Store(key, cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(olCacheTTL),
	})
}

// stats returns current hit/miss counts and resets them.
func (c *olCache) stats() (hits, misses int64) {
	hits = c.hits.Swap(0)
	misses = c.misses.Swap(0)
	return
}

// evictExpired removes entries that have passed their TTL.
func (c *olCache) evictExpired() {
	now := time.Now()
	c.entries.Range(func(key, val any) bool {
		if now.After(val.(cacheEntry).expiresAt) {
			c.entries.Delete(key)
		}
		return true
	})
}

var (
	globalOLClient *cachedOLClient
	olClientOnce   sync.Once
)

// cachedOLClient wraps olClient with a response cache.
type cachedOLClient struct {
	httpClient *http.Client
	baseURL    string
	cache      *olCache
}

// newOLClient returns a singleton OL client with response caching.
// The first call starts a background goroutine that logs cache stats
// and evicts expired entries every hour.
func newOLClient() *cachedOLClient {
	olClientOnce.Do(func() {
		c := &olCache{}
		globalOLClient = &cachedOLClient{
			httpClient: &http.Client{Timeout: 10 * time.Second},
			baseURL:    "https://openlibrary.org",
			cache:      c,
		}
		go func() {
			ticker := time.NewTicker(1 * time.Hour)
			defer ticker.Stop()
			for range ticker.C {
				c.evictExpired()
				hits, misses := c.stats()
				total := hits + misses
				if total > 0 {
					log.Printf("[OL Cache] hits=%d misses=%d total=%d hit_rate=%.1f%%",
						hits, misses, total, float64(hits)/float64(total)*100)
				}
			}
		}()
	})
	return globalOLClient
}

func (c *cachedOLClient) get(path string) (map[string]any, error) {
	raw, err := c.getRaw(path)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (c *cachedOLClient) getRaw(path string) ([]byte, error) {
	url := c.baseURL + path

	if cached, ok := c.cache.get(url); ok {
		return cached, nil
	}

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OL API returned %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	c.cache.set(url, body)
	return body, nil
}
