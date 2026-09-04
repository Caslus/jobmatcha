package scanner

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimitCoordinator serializes provider cooldown state without changing the
// execution timeout configured on the wrapped HTTP client.
type RateLimitCoordinator struct {
	client         *http.Client
	now            func() time.Time
	fallbackLimits map[string]requestLimit

	mu         sync.Mutex
	cooldowns  map[string]time.Time
	dispatches map[string][]time.Time
}

type requestLimit struct {
	maxRequests int
	window      time.Duration
}

// NewRateLimitCoordinator creates a request coordinator for one scanner engine.
func NewRateLimitCoordinator(client *http.Client) *RateLimitCoordinator {
	return &RateLimitCoordinator{
		client: client,
		now:    time.Now,
		fallbackLimits: map[string]requestLimit{
			"workable": {maxRequests: 10, window: 10 * time.Second},
		},
		cooldowns:  make(map[string]time.Time),
		dispatches: make(map[string][]time.Time),
	}
}

// Client returns a provider-specific request executor backed by this coordinator.
func (c *RateLimitCoordinator) Client(provider string) *ProviderHTTPClient {
	return &ProviderHTTPClient{coordinator: c, provider: provider}
}

// ProviderHTTPClient applies cooldowns for a single provider.
type ProviderHTTPClient struct {
	coordinator *RateLimitCoordinator
	provider    string
}

// Do waits for the provider cooldown, dispatches the request, and records a
// later cooldown advertised by the response. A rate-limited request with a
// usable cooldown is retried once.
func (c *ProviderHTTPClient) Do(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		if err := c.coordinator.wait(req.Context(), c.provider); err != nil {
			return nil, err
		}
		if err := req.Context().Err(); err != nil {
			return nil, err
		}

		resp, err := c.coordinator.client.Do(req)
		if err != nil || resp == nil {
			return resp, err
		}
		hasCooldown := c.coordinator.observe(c.provider, resp.Header)
		if resp.StatusCode != http.StatusTooManyRequests || !hasCooldown || attempt == 1 {
			return resp, nil
		}
		resp.Body.Close()
	}
}

func (c *RateLimitCoordinator) wait(ctx context.Context, provider string) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		now := c.now()
		c.mu.Lock()
		cooldown := c.cooldowns[provider]
		if cooldown.After(now) {
			c.mu.Unlock()
			if err := waitFor(ctx, cooldown.Sub(now)); err != nil {
				return err
			}
			continue
		}

		limit, limited := c.fallbackLimits[provider]
		if !limited {
			c.mu.Unlock()
			return nil
		}

		dispatches := c.dispatches[provider]
		cutoff := now.Add(-limit.window)
		for len(dispatches) > 0 && !dispatches[0].After(cutoff) {
			dispatches = dispatches[1:]
		}
		if len(dispatches) < limit.maxRequests {
			c.dispatches[provider] = append(dispatches, now)
			c.mu.Unlock()
			return nil
		}

		delay := dispatches[0].Add(limit.window).Sub(now)
		c.dispatches[provider] = dispatches
		c.mu.Unlock()
		if err := waitFor(ctx, delay); err != nil {
			return err
		}
	}
}

func waitFor(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (c *RateLimitCoordinator) observe(provider string, header http.Header) bool {
	until, ok := cooldownFromHeaders(header, c.now())
	if !ok {
		return false
	}
	c.mu.Lock()
	if until.After(c.cooldowns[provider]) {
		c.cooldowns[provider] = until
	}
	c.mu.Unlock()
	return true
}

func cooldownFromHeaders(header http.Header, now time.Time) (time.Time, bool) {
	if retryAfter, ok := parseRetryAfter(headerValue(header, "Retry-After"), now); ok {
		return retryAfter, true
	}
	remaining, err := strconv.ParseInt(strings.TrimSpace(headerValue(header, "X-RateLimit-Remaining")), 10, 64)
	if err != nil || remaining > 0 {
		return time.Time{}, false
	}
	return parseRateLimitReset(headerValue(header, "X-RateLimit-Reset"), now)
}

func headerValue(header http.Header, name string) string {
	for key, values := range header {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func parseRetryAfter(value string, now time.Time) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds >= 0 {
		until := now.Add(time.Duration(seconds) * time.Second)
		return until, true
	}
	if until, err := http.ParseTime(value); err == nil && until.After(now) {
		return until, true
	}
	return time.Time{}, false
}

func parseRateLimitReset(value string, now time.Time) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil {
		until := time.Unix(seconds, 0)
		return until, until.After(now)
	}
	if until, err := http.ParseTime(value); err == nil && until.After(now) {
		return until, true
	}
	return time.Time{}, false
}
