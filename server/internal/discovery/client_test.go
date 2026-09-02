package discovery

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

type roundTripper func(*http.Request) (*http.Response, error)

func (f roundTripper) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func resolverFor(ip string) Resolver {
	return func(context.Context, string) ([]net.IPAddr, error) { return []net.IPAddr{{IP: net.ParseIP(ip)}}, nil }
}

func response(req *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}
}

func TestHTTPClientRejectsPrivateAndLoopbackDestinations(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "10.0.0.1", "169.254.169.254", "::1"} {
		client := NewHTTPClient(HTTPClientOptions{Resolver: resolverFor(ip)})
		if _, err := client.Fetch(context.Background(), "https://careers.example.test/"); err == nil || !strings.Contains(err.Error(), "non-public") {
			t.Fatalf("%s: %v", ip, err)
		}
	}
}

func TestHTTPClientRejectsPrivateRedirectAndLargeOrTimedOutResponses(t *testing.T) {
	resolver := func(_ context.Context, host string) ([]net.IPAddr, error) {
		if host == "private.test" {
			return []net.IPAddr{{IP: net.ParseIP("127.0.0.1")}}, nil
		}
		return []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, nil
	}
	redirectClient := NewHTTPClient(HTTPClientOptions{Resolver: resolver, Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusFound, Header: http.Header{"Location": {"http://private.test/"}}, Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
	})})
	if _, err := redirectClient.Fetch(context.Background(), "https://public.test/"); err == nil || !strings.Contains(err.Error(), "non-public") {
		t.Fatalf("redirect error = %v", err)
	}
	largeClient := NewHTTPClient(HTTPClientOptions{Resolver: resolverFor("93.184.216.34"), MaxBodyBytes: 4, Transport: roundTripper(func(req *http.Request) (*http.Response, error) { return response(req, http.StatusOK, "12345"), nil })})
	if _, err := largeClient.Fetch(context.Background(), "https://public.test/"); err != ErrResponseTooLarge {
		t.Fatalf("large body error = %v", err)
	}
	timeoutClient := NewHTTPClient(HTTPClientOptions{Resolver: resolverFor("93.184.216.34"), Timeout: time.Millisecond, Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		<-req.Context().Done()
		return nil, req.Context().Err()
	})})
	if _, err := timeoutClient.Fetch(context.Background(), "https://public.test/"); err == nil {
		t.Fatal("expected timeout")
	}
}

func TestDialValidatedAddressesTriesEachPublicAddress(t *testing.T) {
	first := net.ParseIP("93.184.216.34")
	second := net.ParseIP("93.184.216.35")
	resolver := func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: first}, {IP: second}}, nil
	}
	firstErr := errors.New("first edge unavailable")
	dialed := []string{}
	client, server := net.Pipe()
	defer server.Close()
	connection, err := dialValidatedAddresses(context.Background(), "tcp", "cdn.example:443", resolver, func(_ context.Context, _ string, address string) (net.Conn, error) {
		dialed = append(dialed, address)
		if address == net.JoinHostPort(first.String(), "443") {
			return nil, firstErr
		}
		return client, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	if got, want := dialed, []string{"93.184.216.34:443", "93.184.216.35:443"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("dialed = %#v, want %#v", got, want)
	}
}
