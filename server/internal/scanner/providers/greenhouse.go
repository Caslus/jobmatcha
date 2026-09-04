package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/caslus/jobmatcha/internal/model"
)

type greenhouseJob struct {
	ID          json.Number `json:"id"`
	Title       string      `json:"title"`
	AbsoluteURL string      `json:"absolute_url"`
	Content     string      `json:"content"`
	UpdatedAt   string      `json:"updated_at"`
	Location    *struct {
		Name string `json:"name"`
	} `json:"location"`
	Departments []struct {
		Name string `json:"name"`
	} `json:"departments"`
}

type greenhouseResponse struct {
	Jobs []greenhouseJob `json:"jobs"`
}

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type Greenhouse struct {
	HTTPClient HTTPClient
}

func (p *Greenhouse) Name() string { return "greenhouse" }

func (p *Greenhouse) RecognizeBoard(u *url.URL) (model.BoardIdentity, bool) {
	host := strings.ToLower(u.Hostname())
	if (host != "boards.greenhouse.io" && host != "job-boards.greenhouse.io") || (u.Scheme != "http" && u.Scheme != "https") {
		return model.BoardIdentity{}, false
	}
	slug := strings.Split(strings.Trim(u.EscapedPath(), "/"), "/")[0]
	if slug == "" {
		return model.BoardIdentity{}, false
	}
	decoded, err := url.PathUnescape(slug)
	if err != nil || strings.Contains(decoded, "/") {
		return model.BoardIdentity{}, false
	}
	return model.BoardIdentity{Provider: p.Name(), BoardIdentifier: strings.ToLower(decoded), CanonicalURL: "https://boards.greenhouse.io/" + url.PathEscape(strings.ToLower(decoded))}, true
}

func (p *Greenhouse) ValidateBoard(ctx context.Context, board model.BoardIdentity) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, board.CanonicalURL, nil)
	if err != nil {
		return fmt.Errorf("greenhouse %s: build validation request: %w", board.BoardIdentifier, err)
	}
	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("greenhouse %s: validate: %w", board.BoardIdentifier, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("greenhouse %s: validation HTTP %d", board.BoardIdentifier, resp.StatusCode)
	}
	return nil
}

func (p *Greenhouse) Fetch(ctx context.Context, company *model.Company, board *model.CareerBoard) ([]*model.Role, error) {
	slug := board.BoardIdentifier
	if slug == "" {
		return nil, fmt.Errorf("greenhouse: no slug for company %s", company.Name)
	}

	apiURL := fmt.Sprintf("https://boards-api.greenhouse.io/v1/boards/%s/jobs?content=true", slug)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("greenhouse %s: build request: %w", slug, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("greenhouse %s: fetch: %w", slug, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("greenhouse %s: HTTP %d", slug, resp.StatusCode)
	}

	var data greenhouseResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("greenhouse %s: decode: %w", slug, err)
	}

	roles := make([]*model.Role, 0, len(data.Jobs))
	for _, j := range data.Jobs {
		locName := ""
		if j.Location != nil {
			locName = j.Location.Name
		}

		dept := ""
		if len(j.Departments) > 0 {
			dept = j.Departments[0].Name
		}

		// Store raw HTML content from Greenhouse — it's already structured.
		description := j.Content
		descriptionFormat := "html"
		if len(description) > 12000 {
			description = description[:12000]
		}
		postedAt := parseDate(j.UpdatedAt)

		roles = append(roles, &model.Role{
			Title:             strings.TrimSpace(j.Title),
			URL:               j.AbsoluteURL,
			Department:        dept,
			Location:          locName,
			Description:       description,
			DescriptionFormat: descriptionFormat,
			PostedAt:          postedAt,
			Status:            "seen",
		})
	}
	return roles, nil
}
