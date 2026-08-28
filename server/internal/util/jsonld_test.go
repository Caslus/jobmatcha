package util

import "testing"

func TestFindJobPostingJSONLDReturnsFirstDirectJobPosting(t *testing.T) {
	html := []byte(`
		<script type="application/ld+json">{"@context":"https://schema.org","@type":"Organization","name":"Acme"}</script>
		<script type="application/ld+json">{"@context":"https://schema.org","@type":"JobPosting","title":"Backend Engineer","description":"Build APIs","datePosted":"2026-08-20","employmentType":"FULL_TIME","url":"https://example.com/jobs/backend","jobLocation":{"@type":"Place","address":{"addressLocality":"Sao Paulo"}}}</script>
		<script type="application/ld+json">{"@type":"JobPosting","title":"Ignored later role"}</script>
	`)

	job := FindJobPostingJSONLD(html)
	if job == nil {
		t.Fatal("FindJobPostingJSONLD() = nil")
	}
	if job.Title != "Backend Engineer" || job.Description != "Build APIs" || job.DatePosted != "2026-08-20" || job.EmploymentType != "FULL_TIME" || job.URL != "https://example.com/jobs/backend" {
		t.Fatalf("job = %#v", job)
	}
	if string(job.JobLocation) != `{"@type":"Place","address":{"addressLocality":"Sao Paulo"}}` {
		t.Fatalf("JobLocation = %s", job.JobLocation)
	}
}

func TestFindJobPostingJSONLDFindsJobPostingInGraph(t *testing.T) {
	html := []byte(`
		<script type="application/ld+json">not json</script>
		<script type="application/ld+json">{"@context":"https://schema.org","@graph":[{"@type":"Organization","name":"Acme"},{"@type":"JobPosting","title":"Platform Engineer","url":"https://example.com/jobs/platform"}]}</script>
	`)

	job := FindJobPostingJSONLD(html)
	if job == nil {
		t.Fatal("FindJobPostingJSONLD() = nil")
	}
	if job.Type != "JobPosting" || job.Title != "Platform Engineer" || job.URL != "https://example.com/jobs/platform" {
		t.Fatalf("job = %#v", job)
	}
}

func TestFindJobPostingJSONLDReturnsNilWhenNoJobPostingExists(t *testing.T) {
	tests := []struct {
		name string
		html string
	}{
		{name: "no JSON-LD", html: `<html><body>No structured data</body></html>`},
		{name: "invalid JSON-LD", html: `<script type="application/ld+json">{not json}</script>`},
		{name: "non-job JSON-LD", html: `<script type="application/ld+json">{"@type":"Organization","name":"Acme"}</script>`},
		{name: "graph without job", html: `<script type="application/ld+json">{"@graph":[{"@type":"Organization"}]}</script>`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if job := FindJobPostingJSONLD([]byte(tt.html)); job != nil {
				t.Fatalf("FindJobPostingJSONLD() = %#v, want nil", job)
			}
		})
	}
}
