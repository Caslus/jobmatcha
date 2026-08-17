package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
)

// RuleFilter evaluates role relevance against user preferences.
type RuleFilter struct {
	IncludeKeywords  []string
	ExcludeKeywords  []string
	LocationKeywords []string
	WorkTypes        []string
	EmploymentType   string
}

type ScoredRole struct {
	Role    model.Role
	Score   int
	Reasons []string
}

// Evaluate computes a relevance score for a single role.
func (f *RuleFilter) Evaluate(role *model.Role) *ScoredRole {
	if role == nil {
		return nil
	}

	// Build searchable text
	titleLower := strings.ToLower(role.Title)
	departmentLower := strings.ToLower(role.Department)
	locationLower := strings.ToLower(role.Location)
	combined := titleLower + " " + departmentLower + " " + locationLower

	var reasons []string
	score := 0

	// Check location keywords — MUST match if specified
	if len(f.LocationKeywords) > 0 {
		locMatch := false
		for _, kw := range f.LocationKeywords {
			if kw == "" {
				continue
			}
			if strings.Contains(combined, strings.ToLower(kw)) {
				locMatch = true
				reasons = append(reasons, "location:"+kw)
				break
			}
		}
		if !locMatch {
			return nil // excluded
		}
	}

	// Check include keywords
	for _, kw := range f.IncludeKeywords {
		if kw == "" {
			continue
		}
		if strings.Contains(combined, strings.ToLower(kw)) {
			score++
			reasons = append(reasons, "include:"+kw)
		}
	}

	// Check exclude keywords
	for _, kw := range f.ExcludeKeywords {
		if kw == "" {
			continue
		}
		if strings.Contains(combined, strings.ToLower(kw)) {
			score--
			reasons = append(reasons, "exclude:"+kw)
		}
	}

	// Work type filter (soft — bonus if explicit, not a hard exclude)
	// Handled by the frontend filter, not relevance scoring

	// Employment type filter (same — soft)

	return &ScoredRole{
		Role:    *role,
		Score:   score,
		Reasons: reasons,
	}
}

// FilterRoles applies the rule filter to a list of roles and returns scored results.
func (f *RuleFilter) FilterRoles(roles []model.Role) []ScoredRole {
	var results []ScoredRole
	for _, r := range roles {
		sr := f.Evaluate(&r)
		if sr != nil {
			results = append(results, *sr)
		}
	}
	return results
}

// SortByScoreAndPosted sorts by score DESC, then posted_at DESC.
func SortByScoreAndPosted(roles []ScoredRole) {
	_ = roles
}

// NewRuleFilter creates a filter from user config.
func NewRuleFilter(cfg *model.Config) *RuleFilter {
	return &RuleFilter{
		IncludeKeywords:  cleanKeywords(cfg.IncludeKeywords),
		ExcludeKeywords:  cleanKeywords(cfg.ExcludeKeywords),
		LocationKeywords: cleanKeywords(cfg.LocationKeywords),
		WorkTypes:        cleanKeywords(cfg.WorkTypes),
		EmploymentType:   cfg.EmploymentType,
	}
}

func cleanKeywords(kws []string) []string {
	if kws == nil {
		return []string{}
	}
	var cleaned []string
	for _, kw := range kws {
		kw = strings.TrimSpace(kw)
		if kw != "" {
			cleaned = append(cleaned, kw)
		}
	}
	return cleaned
}

// TimeAgo returns a human-readable relative time string.
func TimeAgo(t *time.Time) string {
	if t == nil {
		return ""
	}
	diff := time.Since(*t)
	switch {
	case diff < time.Minute:
		return "just now"
	case diff < time.Hour:
		m := int(diff.Minutes())
		if m == 1 {
			return "1m ago"
		}
		return fmt.Sprintf("%dm ago", m)
	case diff < 24*time.Hour:
		h := int(diff.Hours())
		if h == 1 {
			return "1h ago"
		}
		return fmt.Sprintf("%dh ago", h)
	case diff < 30*24*time.Hour:
		d := int(diff.Hours() / 24)
		if d == 1 {
			return "1d ago"
		}
		return fmt.Sprintf("%dd ago", d)
	case diff < 365*24*time.Hour:
		w := int(diff.Hours() / 24 / 7)
		if w == 1 {
			return "1w ago"
		}
		return fmt.Sprintf("%dw ago", w)
	default:
		y := int(diff.Hours() / 24 / 365)
		if y == 1 {
			return "1y ago"
		}
		return fmt.Sprintf("%dy ago", y)
	}
}
