// Package discovery contains the bounded, non-persistent careers-site
// discovery workflow.
package discovery

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout      = 12 * time.Second
	defaultMaxRedirects = 3
	defaultMaxBodyBytes = int64(2 << 20)
)

var ErrDisallowedDestination = errors.New("discovery URL resolves to a non-public destination")
var ErrResponseTooLarge = errors.New("discovery response exceeds the size limit")

type Resolver func(context.Context, string) ([]net.IPAddr, error)

type HTTPClientOptions struct {
	Timeout      time.Duration
	MaxRedirects int
	MaxBodyBytes int64
	Resolver     Resolver
	Transport    http.RoundTripper // test seam; production uses a validated dialer
}

type Page struct {
	URL         *url.URL
	ContentType string
	Body        []byte
}

// HTTPClient makes discovery-only requests. It deliberately does not inherit
// proxy settings and resolves each destination to a vetted IP before dialing.
type HTTPClient struct {
	client       *http.Client
	resolver     Resolver
	maxBodyBytes int64
}

func NewHTTPClient(options HTTPClientOptions) *HTTPClient {
	if options.Timeout <= 0 {
		options.Timeout = defaultTimeout
	}
	if options.MaxRedirects <= 0 {
		options.MaxRedirects = defaultMaxRedirects
	}
	if options.MaxBodyBytes <= 0 {
		options.MaxBodyBytes = defaultMaxBodyBytes
	}
	if options.Resolver == nil {
		options.Resolver = defaultResolver
	}
	transport := options.Transport
	if transport == nil {
		transport = &http.Transport{
			Proxy:                 nil,
			DialContext:           validatedDialer(options.Resolver),
			ForceAttemptHTTP2:     true,
			TLSHandshakeTimeout:   5 * time.Second,
			ResponseHeaderTimeout: 8 * time.Second,
			IdleConnTimeout:       15 * time.Second,
		}
	}
	c := &HTTPClient{resolver: options.Resolver, maxBodyBytes: options.MaxBodyBytes}
	c.client = &http.Client{Timeout: options.Timeout, Transport: transport, CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= options.MaxRedirects {
			return http.ErrUseLastResponse
		}
		return c.validateURL(req.Context(), req.URL)
	}}
	return c
}

func defaultResolver(ctx context.Context, host string) ([]net.IPAddr, error) {
	addresses, err := net.DefaultResolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	result := make([]net.IPAddr, len(addresses))
	for i, address := range addresses {
		result[i] = net.IPAddr{IP: net.IP(address.AsSlice())}
	}
	return result, nil
}

func (c *HTTPClient) Fetch(ctx context.Context, raw string) (*Page, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return nil, fmt.Errorf("parse discovery URL: %w", err)
	}
	if err := c.validateURL(ctx, u); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("build discovery request: %w", err)
	}
	req.Header.Set("Accept", "text/html,application/xhtml+xml")
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch discovery URL: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("fetch discovery URL: HTTP %d", resp.StatusCode)
	}
	limited := io.LimitReader(resp.Body, c.maxBodyBytes+1)
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read discovery response: %w", err)
	}
	if int64(len(body)) > c.maxBodyBytes {
		return nil, ErrResponseTooLarge
	}
	return &Page{URL: resp.Request.URL, ContentType: resp.Header.Get("Content-Type"), Body: body}, nil
}

func (c *HTTPClient) validateURL(ctx context.Context, u *url.URL) error {
	if u == nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return fmt.Errorf("%w: only absolute HTTP(S) URLs are allowed", ErrDisallowedDestination)
	}
	addresses, err := c.resolver(ctx, u.Hostname())
	if err != nil {
		return fmt.Errorf("resolve discovery host %q: %w", u.Hostname(), err)
	}
	if len(addresses) == 0 {
		return fmt.Errorf("resolve discovery host %q: no addresses", u.Hostname())
	}
	for _, address := range addresses {
		if !isPublicIP(address.IP) {
			return fmt.Errorf("%w: %s", ErrDisallowedDestination, address.IP)
		}
	}
	return nil
}

func validatedDialer(resolve Resolver) func(context.Context, string, string) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: 5 * time.Second}
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		return dialValidatedAddresses(ctx, network, address, resolve, dialer.DialContext)
	}
}

// dialValidatedAddresses preserves the DNS-rebinding protection of resolving
// before dialing, but tries every vetted address. CDN hostnames commonly have
// multiple edges and an individual edge can be temporarily unreachable.
func dialValidatedAddresses(ctx context.Context, network, address string, resolve Resolver, dial func(context.Context, string, string) (net.Conn, error)) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}
	addresses, err := resolve(ctx, host)
	if err != nil {
		return nil, err
	}
	if len(addresses) == 0 {
		return nil, fmt.Errorf("resolve discovery host %q: no addresses", host)
	}
	for _, candidate := range addresses {
		if !isPublicIP(candidate.IP) {
			return nil, fmt.Errorf("%w: %s", ErrDisallowedDestination, candidate.IP)
		}
	}
	var failures []error
	for _, candidate := range addresses {
		connection, err := dial(ctx, network, net.JoinHostPort(candidate.IP.String(), port))
		if err == nil {
			return connection, nil
		}
		failures = append(failures, err)
	}
	return nil, fmt.Errorf("connect discovery host %q: %w", host, errors.Join(failures...))
}

func isPublicIP(ip net.IP) bool {
	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return false
	}
	addr = addr.Unmap()
	if !addr.IsGlobalUnicast() || addr.IsLoopback() || addr.IsLinkLocalUnicast() || addr.IsPrivate() {
		return false
	}
	if addr.Is4() {
		value := addr.As4()
		if value[0] == 0 || value[0] >= 224 || (value[0] == 100 && value[1]&0xc0 == 0x40) || (value[0] == 192 && value[1] == 0 && value[2] == 0) {
			return false
		}
	}
	return true
}

func IsHTML(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(strings.Split(contentType, ";")[0]))
	return contentType == "text/html" || contentType == "application/xhtml+xml" || contentType == ""
}
