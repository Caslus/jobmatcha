package providers

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (f fixtureTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func fixtureBody(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("read fixture %s: %v", name, err)
	}
	return body
}

func fixtureResponse(req *http.Request, status int, body []byte) *http.Response {
	return &http.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(string(body))),
		Header:     make(http.Header),
		Request:    req,
	}
}

func TestGreenhouseFetchMapsFixtureJobs(t *testing.T) {
	jobs := fixtureBody(t, "greenhouse_jobs.json")
	client := &http.Client{Transport: fixtureTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://boards-api.greenhouse.io/v1/boards/acme/jobs?content=true" {
			t.Fatalf("unexpected request: %s", req.URL)
		}
		return fixtureResponse(req, http.StatusOK, jobs), nil
	})}

	roles, err := (&Greenhouse{HTTPClient: client}).Fetch(context.Background(), &model.Company{Name: "Acme"}, &model.CareerBoard{BoardIdentifier: "acme"})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(roles) != 2 {
		t.Fatalf("len(roles) = %d, want 2", len(roles))
	}

	role := roles[0]
	if role.Title != "Staff Platform Engineer" || role.URL != "https://boards.greenhouse.io/acme/jobs/123" || role.Department != "Engineering" || role.Location != "Remote, Brazil" {
		t.Fatalf("mapped role = %#v", role)
	}
	if role.Description != "<p>Build the platform.</p>" || role.DescriptionFormat != "html" || role.Status != "seen" {
		t.Fatalf("description mapping = %#v", role)
	}
	wantPostedAt := time.Date(2026, time.August, 20, 10, 30, 0, 0, time.UTC)
	if role.PostedAt == nil || !role.PostedAt.Equal(wantPostedAt) {
		t.Fatalf("PostedAt = %v, want %v", role.PostedAt, wantPostedAt)
	}
	if roles[1].Department != "" || roles[1].Location != "" || roles[1].PostedAt != nil {
		t.Fatalf("optional field fallback = %#v", roles[1])
	}
}

func TestGreenhouseRecognizesEquivalentBoardURLsAndValidates(t *testing.T) {
	provider := &Greenhouse{HTTPClient: &http.Client{Transport: fixtureTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://boards.greenhouse.io/acme" {
			t.Fatalf("validation request = %s", req.URL)
		}
		return fixtureResponse(req, http.StatusOK, nil), nil
	})}}
	for _, raw := range []string{"https://boards.greenhouse.io/Acme/jobs/1", "https://job-boards.greenhouse.io/acme/"} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		identity, ok := provider.RecognizeBoard(u)
		if !ok || identity.BoardIdentifier != "acme" || identity.CanonicalURL != "https://boards.greenhouse.io/acme" {
			t.Fatalf("recognition for %s = %#v, %v", raw, identity, ok)
		}
		if err := provider.ValidateBoard(context.Background(), identity); err != nil {
			t.Fatalf("validate %s: %v", raw, err)
		}
	}
	for _, raw := range []string{"https://example.test/acme", "https://boards.greenhouse.io/"} {
		u, _ := url.Parse(raw)
		if _, ok := provider.RecognizeBoard(u); ok {
			t.Fatalf("unexpected match for %s", raw)
		}
	}
	failing := &Greenhouse{HTTPClient: &http.Client{Transport: fixtureTransport(func(req *http.Request) (*http.Response, error) {
		return fixtureResponse(req, http.StatusNotFound, nil), nil
	})}}
	if err := failing.ValidateBoard(context.Background(), model.BoardIdentity{Provider: "greenhouse", BoardIdentifier: "missing", CanonicalURL: "https://boards.greenhouse.io/missing"}); err == nil {
		t.Fatal("expected validation failure")
	}
}

func TestWorkableFetchFallsBackToWidgetDescriptionAndMergesLocations(t *testing.T) {
	jobs := fixtureBody(t, "workable_jobs.json")
	client := &http.Client{Transport: fixtureTransport(func(req *http.Request) (*http.Response, error) {
		switch req.URL.String() {
		case "https://apply.workable.com/api/v1/widget/accounts/acme":
			return fixtureResponse(req, http.StatusOK, jobs), nil
		case "https://apply.workable.com/acme/jobs/view/backend-engineer.md":
			return fixtureResponse(req, http.StatusNotFound, nil), nil
		default:
			t.Fatalf("unexpected request: %s", req.URL)
			return nil, nil
		}
	})}

	roles, err := (&Workable{HTTPClient: client}).Fetch(context.Background(), &model.Company{Name: "Acme"}, &model.CareerBoard{BoardIdentifier: "acme"})
	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(roles) != 1 {
		t.Fatalf("len(roles) = %d, want 1", len(roles))
	}

	role := roles[0]
	if role.Title != "Backend Engineer" || role.Department != "Engineering" || role.URL != "https://apply.workable.com/acme/j/backend-engineer/" {
		t.Fatalf("mapped role = %#v", role)
	}
	if role.Location != "Sao Paulo, SP, Brazil; Lisbon, Portugal" {
		t.Fatalf("Location = %q", role.Location)
	}
	if role.Description != "Build reliable services." || role.DescriptionFormat != "plain" {
		t.Fatalf("widget fallback = %#v", role)
	}
	wantPostedAt := time.Date(2026, time.August, 21, 0, 0, 0, 0, time.UTC)
	if role.PostedAt == nil || !role.PostedAt.Equal(wantPostedAt) {
		t.Fatalf("PostedAt = %v, want %v", role.PostedAt, wantPostedAt)
	}
}

func TestWorkableRecognizesAndValidatesBoards(t *testing.T) {
	provider := &Workable{HTTPClient: &http.Client{Transport: fixtureTransport(func(req *http.Request) (*http.Response, error) {
		if req.URL.String() != "https://apply.workable.com/api/v1/widget/accounts/acme" {
			t.Fatalf("validation request = %s", req.URL)
		}
		return fixtureResponse(req, http.StatusOK, nil), nil
	})}}
	for _, raw := range []string{"https://apply.workable.com/Acme/", "http://apply.workable.com/acme/jobs/view/123"} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		identity, ok := provider.RecognizeBoard(u)
		if !ok || identity.BoardIdentifier != "acme" || identity.CanonicalURL != "https://apply.workable.com/acme/" {
			t.Fatalf("recognition for %s = %#v, %v", raw, identity, ok)
		}
		if err := provider.ValidateBoard(context.Background(), identity); err != nil {
			t.Fatalf("validate %s: %v", raw, err)
		}
	}
	for _, raw := range []string{"https://jobs.workable.com/acme", "https://apply.workable.com/"} {
		u, _ := url.Parse(raw)
		if _, ok := provider.RecognizeBoard(u); ok {
			t.Fatalf("unexpected match for %s", raw)
		}
	}
	failing := &Workable{HTTPClient: &http.Client{Transport: fixtureTransport(func(req *http.Request) (*http.Response, error) {
		return fixtureResponse(req, http.StatusBadGateway, nil), nil
	})}}
	if err := failing.ValidateBoard(context.Background(), model.BoardIdentity{Provider: "workable", BoardIdentifier: "missing"}); err == nil {
		t.Fatal("expected validation failure")
	}
}
