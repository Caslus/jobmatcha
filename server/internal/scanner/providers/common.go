package providers

import (
	"encoding/json"
	"regexp"
	"strings"
	"time"
)

var (
	htmlTagRe    = regexp.MustCompile(`<[^>]+>`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

// StripHTML removes HTML tags and collapses whitespace.
// Truncates to maxLen characters (default 12000 if maxLen <= 0).
func StripHTML(html string, maxLen int) string {
	if maxLen <= 0 {
		maxLen = 12000
	}
	text := htmlTagRe.ReplaceAllString(html, " ")
	text = whitespaceRe.ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)
	if len(text) > maxLen {
		text = text[:maxLen]
	}
	return text
}

// parseDate attempts to parse a date string in various common formats.
func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	formats := []string{
		time.RFC3339,
		"2006-01-02",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05-07:00",
		"2006/01/02",
		"January 2, 2006",
		"2006-01-02 15:04:05",
	}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return &t
		}
	}
	return nil
}

// extractLocation parses a JSON-LD jobLocation field into a human-readable string.
func extractLocation(raw []byte) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}

	// Try structured address first
	var addr struct {
		Address struct {
			AddressLocality string `json:"addressLocality"`
			AddressRegion   string `json:"addressRegion"`
			AddressCountry  any    `json:"addressCountry"`
		} `json:"address"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(raw, &addr); err == nil {
		if addr.Name != "" {
			return addr.Name
		}
		if addr.Address.AddressLocality != "" {
			parts := []string{addr.Address.AddressLocality}
			if addr.Address.AddressRegion != "" {
				parts = append(parts, addr.Address.AddressRegion)
			}
			if c, ok := addr.Address.AddressCountry.(string); ok && c != "" {
				parts = append(parts, c)
			}
			return strings.Join(parts, ", ")
		}
	}

	// Fallback: try plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return s
	}

	return ""
}
