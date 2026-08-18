package service

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
)

// keywordMatcher matches a keyword against lowercase text.
// Uses word boundaries for latin keywords; plain substring for CJK
// (Go's \b does not work around Japanese characters).
// Word boundaries are only applied when the keyword starts/ends with
// a word character, so keywords like "c#" or "full-stack" match correctly.
type keywordMatcher struct {
	raw string
	re  *regexp.Regexp // nil when the keyword contains CJK or no boundary
}

func newKeywordMatcher(kw string) keywordMatcher {
	kw = strings.ToLower(strings.TrimSpace(kw))
	if kw == "" {
		return keywordMatcher{raw: ""}
	}
	if containsCJK(kw) {
		return keywordMatcher{raw: kw}
	}
	pattern := regexp.QuoteMeta(kw)
	runes := []rune(kw)
	if isWordRune(runes[0]) {
		pattern = `\b` + pattern
	}
	if isWordRune(runes[len(runes)-1]) {
		pattern = pattern + `\b`
	}
	re := regexp.MustCompile(pattern)
	return keywordMatcher{raw: kw, re: re}
}

func (m keywordMatcher) Match(text string) bool {
	if m.raw == "" {
		return false
	}
	if m.re != nil {
		return m.re.MatchString(text)
	}
	return strings.Contains(text, m.raw)
}

// isWordRune matches \w characters (letters, digits, underscore).
func isWordRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_'
}

// containsCJK reports whether s contains hiragana, katakana, or CJK ideographs.
func containsCJK(s string) bool {
	for _, r := range s {
		switch {
		case r >= 0x3040 && r <= 0x30FF, // hiragana + katakana
			r >= 0x4E00 && r <= 0x9FFF: // CJK unified ideographs
			return true
		}
	}
	return false
}

// Scoring weights.
const (
	titleWeight   = 2 // include keyword hit in the title
	fieldWeight   = 1 // include keyword hit in department or location
	workTypeBonus = 1 // role explicitly matches a selected work type
)

// workTypeKeywords maps employment types to text signals (lowercase).
var workTypeKeywords = map[string][]string{
	"full-time":  {"full-time", "full time", "フルタイム", "常勤", "正社員"},
	"part-time":  {"part-time", "part time", "パート", "アルバイト", "非常勤"},
	"contract":   {"contract", "contractor", "契約", "派遣", "業務委託", "フリーランス"},
	"internship": {"intern", "internship", "インターン", "インターンシップ"},
}

// RuleFilter evaluates role relevance against user preferences.
type RuleFilter struct {
	IncludeKeywords  []string
	ExcludeKeywords  []string
	LocationKeywords []string
	WorkTypes        []string
	EmploymentType   string

	includes         []keywordMatcher
	excludes         []keywordMatcher
	locations        []keywordMatcher
	workTypeMatchers map[string][]keywordMatcher
}

// ScoredRole carries the relevance result for a single role.
type ScoredRole struct {
	Role    model.Role
	Score   int
	Percent int
	Reasons []string
}

// NewRuleFilter creates a filter from user config.
func NewRuleFilter(cfg *model.Config) *RuleFilter {
	f := &RuleFilter{
		IncludeKeywords:  cleanKeywords(cfg.IncludeKeywords),
		ExcludeKeywords:  cleanKeywords(cfg.ExcludeKeywords),
		LocationKeywords: cleanKeywords(cfg.LocationKeywords),
		WorkTypes:        cleanKeywords(cfg.WorkTypes),
		EmploymentType:   cfg.EmploymentType,
	}
	f.includes = make([]keywordMatcher, 0, len(f.IncludeKeywords))
	for _, kw := range f.IncludeKeywords {
		f.includes = append(f.includes, newKeywordMatcher(kw))
	}
	f.excludes = make([]keywordMatcher, 0, len(f.ExcludeKeywords))
	for _, kw := range f.ExcludeKeywords {
		f.excludes = append(f.excludes, newKeywordMatcher(kw))
	}
	f.locations = make([]keywordMatcher, 0, len(f.LocationKeywords))
	for _, kw := range f.LocationKeywords {
		f.locations = append(f.locations, newKeywordMatcher(kw))
	}
	f.workTypeMatchers = make(map[string][]keywordMatcher)
	for wt, kws := range workTypeKeywords {
		for _, kw := range kws {
			f.workTypeMatchers[wt] = append(f.workTypeMatchers[wt], newKeywordMatcher(kw))
		}
	}
	return f
}

// Evaluate computes a relevance score for a single role.
// Returns nil when the role must be filtered out (location mismatch,
// exclude keyword hit, or an explicitly advertised non-selected work type).
func (f *RuleFilter) Evaluate(role *model.Role) *ScoredRole {
	if role == nil {
		return nil
	}

	title := strings.ToLower(role.Title)
	dept := strings.ToLower(role.Department)
	loc := strings.ToLower(role.Location)
	combined := title + " " + dept + " " + loc

	var reasons []string

	// 1. Location keywords — MUST match if specified (hard filter)
	if len(f.locations) > 0 {
		matched := false
		for _, m := range f.locations {
			if m.Match(combined) {
				matched = true
				reasons = append(reasons, "location:"+m.raw)
				break
			}
		}
		if !matched {
			return nil
		}
	}

	// 2. Exclude keywords — hard filter (title + department)
	for _, m := range f.excludes {
		if m.Match(title + " " + dept) {
			return nil
		}
	}

	// 3. Work type — drop roles that explicitly advertise a non-selected type.
	// Roles with no work type mention are kept (we can't infer, don't punish).
	var workTypeFound []string
	if len(f.WorkTypes) > 0 {
		workTypeFound = f.detectWorkTypes(combined)
		for _, wt := range workTypeFound {
			if !containsStr(f.WorkTypes, wt) {
				return nil
			}
		}
	}

	// 4. Include keywords — weighted by field
	includeScore := 0
	for _, m := range f.includes {
		switch {
		case m.Match(title):
			includeScore += titleWeight
			reasons = append(reasons, "title:"+m.raw)
		case m.Match(dept):
			includeScore += fieldWeight
			reasons = append(reasons, "dept:"+m.raw)
		case m.Match(loc):
			includeScore += fieldWeight
			reasons = append(reasons, "loc:"+m.raw)
		}
	}

	// 5. Work type bonus — explicit match with a selected type adds signal
	totalScore := includeScore
	if len(workTypeFound) > 0 {
		totalScore += workTypeBonus
		reasons = append(reasons, "work_type:"+strings.Join(workTypeFound, ","))
	}

	// 6. Recency-adjusted score for percentage.
	// Newer roles score higher: a role posted this week keeps full
	// weight, one from 3 months ago gets halved. This spreads the
	// percentage naturally — same include matches but different ages
	// produce different numbers instead of clustering.
	adjusted := float64(includeScore) * recencyFactor(role.PostedAt)

	// 7. Match percentage — saturating scale.
	// score 7+ (adjusted) reaches 100%. Non-round numbers (29, 43, 57…)
	// feel organic.
	percent := 0
	if adjusted > 0 {
		percent = (int(adjusted*100) + 3) / 7
		if percent > 100 {
			percent = 100
		}
	}

	return &ScoredRole{
		Role:    *role,
		Score:   totalScore,
		Percent: percent,
		Reasons: reasons,
	}
}

// detectWorkTypes returns the work types mentioned in text, if any.
func (f *RuleFilter) detectWorkTypes(text string) []string {
	var found []string
	for wt, matchers := range f.workTypeMatchers {
		for _, m := range matchers {
			if m.Match(text) {
				found = append(found, wt)
				break
			}
		}
	}
	return found
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

// SortByScoreAndPosted sorts by percent DESC, then score DESC, then posted_at DESC.
func SortByScoreAndPosted(roles []ScoredRole) {
	_ = roles
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// recencyFactor returns a multiplier for the include score based on
// how recently the role was posted. Uses a linear decay: full weight
// at day 0, smoothly declining to 0.3 at day ~126, then flat.
// This spreads same-score roles naturally across their age.
func recencyFactor(postedAt *time.Time) float64 {
	if postedAt == nil {
		return 0.3
	}
	days := time.Since(*postedAt).Hours() / 24
	factor := 1.0 - days/180.0 // 180-day decay horizon
	if factor < 0.3 {
		return 0.3
	}
	return factor
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