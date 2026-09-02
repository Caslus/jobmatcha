package discovery

import (
	"net/url"
	"regexp"
	"strings"
)

type ExtractedURL struct {
	URL   *url.URL
	Label string
}

var attributeURLPattern = regexp.MustCompile(`(?is)<(?P<tag>a|meta|script)\b[^>]*?(?:href|src|content)\s*=\s*["']([^"']+)["'][^>]*>`)
var anchorPattern = regexp.MustCompile(`(?is)<a\b[^>]*href\s*=\s*["']([^"']+)["'][^>]*>(.*?)</a\s*>`)
var rawURLPattern = regexp.MustCompile(`https?://[^\s"'<>\\]+`)
var tagPattern = regexp.MustCompile(`(?is)<[^>]+>`)
var ogTitleAfterPropertyPattern = regexp.MustCompile(`(?is)<meta\b[^>]*(?:property|name)\s*=\s*["']og:title["'][^>]*content\s*=\s*["']([^"']+)["'][^>]*>`)
var ogTitleBeforePropertyPattern = regexp.MustCompile(`(?is)<meta\b[^>]*content\s*=\s*["']([^"']+)["'][^>]*(?:property|name)\s*=\s*["']og:title["'][^>]*>`)
var titlePattern = regexp.MustCompile(`(?is)<title\b[^>]*>\s*(.*?)\s*</title>`)

// ExtractURLs collects links, URL-bearing metadata and script sources, then
// scans the complete document for absolute URLs in inline/minified scripts.
func ExtractURLs(base *url.URL, html string, limit int) []ExtractedURL {
	if base == nil || limit <= 0 {
		return nil
	}
	urls := make([]ExtractedURL, 0)
	seen := make(map[string]struct{})
	add := func(raw, label string) {
		if len(urls) >= limit {
			return
		}
		u, err := url.Parse(strings.TrimSpace(raw))
		if err != nil {
			return
		}
		u = base.ResolveReference(u)
		if u.Scheme != "http" && u.Scheme != "https" {
			return
		}
		u.Fragment = ""
		key := u.String()
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		urls = append(urls, ExtractedURL{URL: u, Label: strings.TrimSpace(label)})
	}
	for _, match := range anchorPattern.FindAllStringSubmatch(html, -1) {
		add(match[1], tagPattern.ReplaceAllString(match[2], " "))
	}
	for _, match := range attributeURLPattern.FindAllStringSubmatch(html, -1) {
		add(match[2], "")
	}
	for _, raw := range rawURLPattern.FindAllString(html, -1) {
		add(strings.TrimRight(raw, ".,;:)]}"), "")
	}
	return urls
}

func IsRelevantCareerLink(link ExtractedURL) bool {
	text := strings.ToLower(link.Label + " " + link.URL.Path)
	for _, word := range []string{"career", "careers", "recruit", "recruiting", "job", "jobs", "hiring", "work-with-us"} {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}

// EmployerNameSuggestion returns the most useful human-readable employer
// label available from a discovered HTML document.
func EmployerNameSuggestion(html string) string {
	for _, pattern := range []*regexp.Regexp{ogTitleAfterPropertyPattern, ogTitleBeforePropertyPattern} {
		if match := pattern.FindStringSubmatch(html); len(match) > 1 {
			if value := strings.TrimSpace(match[1]); value != "" {
				return value
			}
		}
	}
	if match := titlePattern.FindStringSubmatch(html); len(match) > 1 {
		return strings.TrimSpace(tagPattern.ReplaceAllString(match[1], " "))
	}
	return ""
}
