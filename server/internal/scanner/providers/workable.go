package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/caslus/jobmatcha/internal/model"
)

type workableJob struct {
	Title        string `json:"title"`
	Shortcode    string `json:"shortcode"`
	URL          string `json:"url"`
	Department   string `json:"department"`
	PublishedOn  string `json:"published_on"`
	Description  string `json:"description"`
	Telecommuting bool `json:"telecommuting"`
	CreatedAt    string `json:"created_at"`
	Locations    []struct {
		City    string `json:"city"`
		Region  string `json:"region"`
		Country string `json:"country"`
	} `json:"locations"`
}

type workableResponse struct {
	Name string        `json:"name"`
	Jobs []workableJob `json:"jobs"`
}

type Workable struct {
	HTTPClient *http.Client
}

func (p *Workable) Name() string { return "workable" }

func (p *Workable) Fetch(ctx context.Context, company *model.Company) ([]*model.Role, error) {
	slug := company.ATSSlug
	if slug == "" {
		return nil, fmt.Errorf("workable: no slug for company %s", company.Name)
	}

	apiURL := fmt.Sprintf("https://apply.workable.com/api/v1/widget/accounts/%s", slug)
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("workable %s: build request: %w", slug, err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("workable %s: fetch: %w", slug, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("workable %s: HTTP %d", slug, resp.StatusCode)
	}

	var data workableResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("workable %s: decode: %w", slug, err)
	}

	// Deduplicate by URL — Workable sometimes returns duplicate shortcodes
	// with different locations that should be merged.
	type workableRole struct {
		role     *model.Role
		locParts []string
	}
	byURL := make(map[string]*workableRole)

	for _, j := range data.Jobs {
		uid := j.URL
		if uid == "" {
			continue
		}

		// Merge locations for duplicate shortcodes
		if existing, ok := byURL[uid]; ok {
			for _, loc := range j.Locations {
				parts := strings.TrimSpace(strings.Join([]string{loc.City, loc.Region, loc.Country}, ", "))
				parts = strings.Trim(parts, ", ")
				if parts != "" {
					found := false
					for _, lp := range existing.locParts {
						if lp == parts {
							found = true
							break
						}
					}
					if !found {
						existing.locParts = append(existing.locParts, parts)
					}
				}
			}
			continue
		}

		var locParts []string
		for _, loc := range j.Locations {
			parts := strings.TrimSpace(strings.Join([]string{loc.City, loc.Region, loc.Country}, ", "))
			parts = strings.Trim(parts, ", ")
			if parts != "" {
				locParts = append(locParts, parts)
			}
		}

		role := &model.Role{
			Title:      strings.TrimSpace(j.Title),
			URL:        uid,
			Department: j.Department,
			Location:   strings.Join(locParts, "; "),
			PostedAt:   parseDate(j.PublishedOn),
			Status:     "seen",
		}

		// Fetch markdown description from Workable's .md endpoint
		if j.Shortcode != "" {
			mdURL := fmt.Sprintf("https://apply.workable.com/%s/jobs/view/%s.md", slug, j.Shortcode)
			mdReq, err := http.NewRequestWithContext(ctx, "GET", mdURL, nil)
			if err == nil {
				mdReq.Header.Set("Accept", "text/markdown")
				mdResp, mdErr := p.HTTPClient.Do(mdReq)
				if mdErr == nil && mdResp.StatusCode == http.StatusOK {
					mdBody, _ := io.ReadAll(mdResp.Body)
					role.Description = StripHTML(string(mdBody), 12000)
					mdResp.Body.Close()
				} else if mdResp != nil {
					mdResp.Body.Close()
				}
			}
		}

		// Fallback to widget description
		if role.Description == "" && j.Description != "" {
			role.Description = StripHTML(j.Description, 12000)
		}

		byURL[uid] = &workableRole{role: role, locParts: locParts}
	}

	roles := make([]*model.Role, 0, len(byURL))
	for _, wr := range byURL {
		roles = append(roles, wr.role)
	}
	return roles, nil
}