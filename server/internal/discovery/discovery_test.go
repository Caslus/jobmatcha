package discovery

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/scanner"
)

type testRecognizer struct{}

func (testRecognizer) RecognizeBoard(u *url.URL) (model.BoardIdentity, bool) {
	if u.Host == "boards.test" {
		return model.BoardIdentity{Provider: "test", BoardIdentifier: u.Path[1:], CanonicalURL: "https://boards.test" + u.Path}, true
	}
	return model.BoardIdentity{}, false
}
func (testRecognizer) ValidateBoard(context.Context, model.BoardIdentity) error { return nil }

type testProvider struct {
	testRecognizer
	fail bool
}

func (testProvider) Name() string { return "test" }
func (testProvider) Fetch(context.Context, *model.Company, *model.CareerBoard) ([]*model.Role, error) {
	return nil, nil
}
func (p testProvider) ValidateBoard(context.Context, model.BoardIdentity) error {
	if p.fail {
		return errors.New("missing board")
	}
	return nil
}

type unsupportedRecognizer struct{}

func (unsupportedRecognizer) RecognizeBoard(u *url.URL) (model.BoardIdentity, bool) {
	if u.Host == "unsupported.test" {
		return model.BoardIdentity{Provider: "unsupported", BoardIdentifier: "other", CanonicalURL: "https://unsupported.test/other"}, true
	}
	return model.BoardIdentity{}, false
}
func (unsupportedRecognizer) ValidateBoard(context.Context, model.BoardIdentity) error { return nil }

func TestDiscoverTraversesOnlyRelevantLinksAndCollectsSeparateCandidates(t *testing.T) {
	resolver := func(context.Context, string) ([]net.IPAddr, error) {
		return []net.IPAddr{{IP: net.ParseIP("93.184.216.34")}}, nil
	}
	requests := []string{}
	client := NewHTTPClient(HTTPClientOptions{Resolver: resolver, Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		requests = append(requests, req.URL.String())
		body := `<a href="https://boards.test/paypay">PayPay</a><a href="/group-careers">Group careers</a><a href="https://unrelated.test/about">About</a>`
		if req.URL.Path == "/group-careers" {
			body = `<script>"https://boards.test/paypay-card"</script>`
		}
		return &http.Response{StatusCode: http.StatusOK, Header: http.Header{"Content-Type": {"text/html"}}, Body: io.NopCloser(strings.NewReader(body)), Request: req}, nil
	})})
	registry := scanner.NewRegistry()
	registry.RegisterRecognizer("test", testRecognizer{})
	result, err := NewService(client, registry, Options{MaxPages: 2}).Discover(context.Background(), "https://company.test/careers")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
	if len(requests) != 2 || requests[1] != "https://company.test/group-careers" {
		t.Fatalf("requests = %#v", requests)
	}
}

func TestDiscoverFollowsRelevantGroupLinksBeyondTheFirstTwo(t *testing.T) {
	client := NewHTTPClient(HTTPClientOptions{Resolver: resolverFor("93.184.216.34"), Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/careers" {
			return response(req, http.StatusOK, `<a href="/group-one/careers">Group one careers</a><a href="/group-two/careers">Group two careers</a><a href="/paypay-card/careers">PayPay Card careers</a>`), nil
		}
		if req.URL.Path == "/paypay-card/careers" {
			return response(req, http.StatusOK, `<a href="https://boards.test/paypay-card">Open roles</a>`), nil
		}
		return response(req, http.StatusOK, `<p>No board on this page</p>`), nil
	})})
	registry := scanner.NewRegistry()
	registry.RegisterRecognizer("test", testRecognizer{})

	result, err := NewService(client, registry, Options{MaxPages: 4}).Discover(context.Background(), "https://company.test/careers")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].Identity.BoardIdentifier != "paypay-card" {
		t.Fatalf("candidates = %#v", result.Candidates)
	}
}

func TestDiscoverReportsReadyAndUnsupportedCandidates(t *testing.T) {
	client := NewHTTPClient(HTTPClientOptions{Resolver: resolverFor("93.184.216.34"), Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		return response(req, http.StatusOK, `<a href="https://boards.test/ready">Ready</a><a href="https://unsupported.test/other">Other</a>`), nil
	})})
	registry := scanner.NewRegistry()
	registry.Register(testProvider{})
	registry.RegisterRecognizer("unsupported", unsupportedRecognizer{})
	result, err := NewService(client, registry, Options{}).Discover(context.Background(), "https://company.test/careers")
	if err != nil {
		t.Fatal(err)
	}
	statuses := map[string]string{}
	for _, candidate := range result.Candidates {
		statuses[candidate.Identity.Provider] = candidate.ValidationStatus
	}
	if statuses["test"] != "ready" || statuses["unsupported"] != "unsupported" {
		t.Fatalf("statuses = %#v", statuses)
	}
}

func TestDiscoverReportsInvalidAndIncompleteResults(t *testing.T) {
	client := NewHTTPClient(HTTPClientOptions{Resolver: resolverFor("93.184.216.34"), Transport: roundTripper(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/careers" {
			return response(req, http.StatusOK, `<a href="https://boards.test/missing">Missing</a><a href="/more-careers">More careers</a>`), nil
		}
		return response(req, http.StatusOK, `<a href="https://boards.test/missing">Duplicate</a>`), nil
	})})
	registry := scanner.NewRegistry()
	registry.Register(testProvider{fail: true})
	result, err := NewService(client, registry, Options{MaxPages: 1}).Discover(context.Background(), "https://company.test/careers")
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Candidates) != 1 || result.Candidates[0].ValidationStatus != "invalid" || !result.Incomplete {
		t.Fatalf("result = %#v", result)
	}
}
