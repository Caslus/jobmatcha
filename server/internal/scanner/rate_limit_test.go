package scanner

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func response(req *http.Request, header http.Header) *http.Response {
	return &http.Response{StatusCode: http.StatusOK, Header: header, Body: io.NopCloser(http.NoBody), Request: req}
}

func TestCooldownFromHeaders(t *testing.T) {
	now := time.Date(2026, time.September, 4, 12, 0, 0, 0, time.UTC)
	future := now.Add(2 * time.Minute)

	tests := []struct {
		name   string
		header http.Header
		want   time.Time
		ok     bool
	}{
		{"retry seconds", http.Header{"Retry-After": {"30"}}, now.Add(30 * time.Second), true},
		{"immediate retry", http.Header{"Retry-After": {"0"}}, now, true},
		{"retry date", http.Header{"Retry-After": {future.Format(http.TimeFormat)}}, future, true},
		{"retry wins over reset", http.Header{"Retry-After": {"30"}, "X-RateLimit-Remaining": {"0"}, "X-RateLimit-Reset": {strconv.FormatInt(future.Unix(), 10)}}, now.Add(30 * time.Second), true},
		{"exhausted epoch reset", http.Header{"X-RateLimit-Remaining": {"0"}, "X-RateLimit-Reset": {strconv.FormatInt(future.Unix(), 10)}}, future, true},
		{"exhausted date reset", http.Header{"X-RateLimit-Remaining": {"-1"}, "X-RateLimit-Reset": {future.Format(http.TimeFormat)}}, future, true},
		{"remaining quota", http.Header{"X-RateLimit-Remaining": {"1"}, "X-RateLimit-Reset": {strconv.FormatInt(future.Unix(), 10)}}, time.Time{}, false},
		{"malformed", http.Header{"Retry-After": {"soon"}, "X-RateLimit-Remaining": {"no"}}, time.Time{}, false},
		{"expired", http.Header{"Retry-After": {"-1"}, "X-RateLimit-Remaining": {"0"}, "X-RateLimit-Reset": {strconv.FormatInt(now.Add(-time.Second).Unix(), 10)}}, time.Time{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := cooldownFromHeaders(tt.header, now)
			if ok != tt.ok || (ok && !got.Equal(tt.want)) {
				t.Fatalf("cooldownFromHeaders() = (%v, %v), want (%v, %v)", got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestProviderHTTPClientHonorsCooldownAndIsolation(t *testing.T) {
	var calls atomic.Int32
	coordinator := NewRateLimitCoordinator(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(req, nil), nil
	})})
	coordinator.cooldowns["workable"] = time.Now().Add(60 * time.Millisecond)

	greenhouseRequest, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/greenhouse", nil)
	start := time.Now()
	if _, err := coordinator.Client("greenhouse").Do(greenhouseRequest); err != nil {
		t.Fatalf("greenhouse request: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 40*time.Millisecond {
		t.Fatalf("other provider waited %v for workable cooldown", elapsed)
	}

	workableRequest, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test/workable", nil)
	start = time.Now()
	if _, err := coordinator.Client("workable").Do(workableRequest); err != nil {
		t.Fatalf("workable request: %v", err)
	}
	if elapsed := time.Since(start); elapsed < 45*time.Millisecond {
		t.Fatalf("workable request waited %v, want cooldown delay", elapsed)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("transport calls = %d, want 2", got)
	}
}

func TestProviderHTTPClientCancellationSkipsTransport(t *testing.T) {
	var calls atomic.Int32
	coordinator := NewRateLimitCoordinator(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return response(req, nil), nil
	})})
	coordinator.cooldowns["greenhouse"] = time.Now().Add(time.Second)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.test", nil)

	_, err := coordinator.Client("greenhouse").Do(req)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Do() error = %v, want context cancellation", err)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("transport calls = %d, want 0", got)
	}
}

func TestProviderHTTPClientRetriesRateLimitedRequest(t *testing.T) {
	var calls atomic.Int32
	coordinator := NewRateLimitCoordinator(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if calls.Add(1) == 1 {
			return &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": {"0"}}, Body: io.NopCloser(http.NoBody), Request: req}, nil
		}
		return response(req, nil), nil
	})})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test", nil)

	resp, err := coordinator.Client("workable").Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("transport calls = %d, want 2", got)
	}
}

func TestProviderHTTPClientCancellationStopsRateLimitedRetry(t *testing.T) {
	firstResponse := make(chan struct{})
	var calls atomic.Int32
	coordinator := NewRateLimitCoordinator(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		close(firstResponse)
		return &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": {"1"}}, Body: io.NopCloser(http.NoBody), Request: req}, nil
	})})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.test", nil)
	errCh := make(chan error, 1)
	go func() {
		_, err := coordinator.Client("workable").Do(req)
		errCh <- err
	}()

	<-firstResponse
	cancel()
	if err := <-errCh; !errors.Is(err, context.Canceled) {
		t.Fatalf("Do() error = %v, want context cancellation", err)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("transport calls = %d, want 1", got)
	}
}

func TestProviderHTTPClientLimitsRateLimitedRetries(t *testing.T) {
	var calls atomic.Int32
	coordinator := NewRateLimitCoordinator(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		calls.Add(1)
		return &http.Response{StatusCode: http.StatusTooManyRequests, Header: http.Header{"Retry-After": {"0"}}, Body: io.NopCloser(http.NoBody), Request: req}, nil
	})})
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test", nil)

	resp, err := coordinator.Client("workable").Do(req)
	if err != nil {
		t.Fatalf("Do() error = %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("status = %d, want %d", resp.StatusCode, http.StatusTooManyRequests)
	}
	if got := calls.Load(); got != 2 {
		t.Fatalf("transport calls = %d, want 2", got)
	}
}

func TestProviderHTTPClientRecordsRetryAfterAndConcurrentUpdates(t *testing.T) {
	reset := time.Now().Add(2 * time.Second).Truncate(time.Second)
	coordinator := NewRateLimitCoordinator(&http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return response(req, http.Header{"Retry-After": {"1"}, "X-RateLimit-Remaining": {"0"}, "X-RateLimit-Reset": {strconv.FormatInt(reset.Unix(), 10)}}), nil
	})})
	client := coordinator.Client("workable")

	var group sync.WaitGroup
	for range 10 {
		group.Go(func() {
			req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.test", nil)
			resp, err := client.Do(req)
			if err == nil {
				resp.Body.Close()
			}
		})
	}
	group.Wait()

	coordinator.mu.Lock()
	cooldown := coordinator.cooldowns["workable"]
	coordinator.mu.Unlock()
	if cooldown.Before(time.Now().Add(900 * time.Millisecond)) {
		t.Fatalf("cooldown = %v, want Retry-After delay", cooldown)
	}
}
