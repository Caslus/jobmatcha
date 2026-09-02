package discovery

import (
	"net/url"
	"testing"
)

func TestExtractURLsFindsLinksMetadataScriptSourcesAndMinifiedInlineURLs(t *testing.T) {
	base, _ := url.Parse("https://careers.example.test/company/")
	html := `<a href="/jobs">Careers</a><meta content="https://boards.greenhouse.io/acme"><script src="/assets/app.js"></script><script>window.x="https://apply.workable.com/acme/jobs"</script>`
	found := ExtractURLs(base, html, 20)
	got := map[string]ExtractedURL{}
	for _, item := range found {
		got[item.URL.String()] = item
	}
	for _, want := range []string{"https://careers.example.test/jobs", "https://boards.greenhouse.io/acme", "https://careers.example.test/assets/app.js", "https://apply.workable.com/acme/jobs"} {
		if _, ok := got[want]; !ok {
			t.Errorf("missing %s in %#v", want, found)
		}
	}
	if !IsRelevantCareerLink(got["https://careers.example.test/jobs"]) {
		t.Fatal("career link not relevant")
	}
}

func TestEmployerNameSuggestionPrefersOpenGraphTitle(t *testing.T) {
	html := `<title>Fallback title</title><meta content="PayPay Careers" property="og:title">`
	if got := EmployerNameSuggestion(html); got != "PayPay Careers" {
		t.Fatalf("EmployerNameSuggestion() = %q", got)
	}
}
