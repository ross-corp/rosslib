// Package olhttp provides a rate-limited HTTP client for Open Library API requests.
// All outbound requests to openlibrary.org should go through this client to avoid
// being banned for excessive traffic.
package olhttp

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"
)

// Default rate limits for Open Library requests.
const (
	// DefaultRate is the steady-state requests per second to OL.
	DefaultRate = 5.0
	// DefaultBurst allows short spikes (e.g. the 3-8 concurrent requests
	// that a single book detail page makes).
	DefaultBurst = 15
)

// rateLimitedTransport wraps an http.RoundTripper with a token-bucket limiter.
type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

func (t *rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	if err := t.limiter.Wait(ctx); err != nil {
		return nil, err
	}
	return t.base.RoundTrip(req)
}

// NewClient returns an *http.Client that rate-limits outbound requests using a
// token-bucket algorithm. rps is requests-per-second and burst is the maximum
// number of requests that can fire at once before throttling kicks in.
func NewClient(rps float64, burst int) *http.Client {
	return &http.Client{
		Timeout: 15 * time.Second,
		Transport: &rateLimitedTransport{
			base:    http.DefaultTransport,
			limiter: rate.NewLimiter(rate.Limit(rps), burst),
		},
	}
}

// DefaultClient returns an *http.Client with the default rate limits for
// Open Library traffic.
func DefaultClient() *http.Client {
	return NewClient(DefaultRate, DefaultBurst)
}
