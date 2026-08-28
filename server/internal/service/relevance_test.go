package service

import (
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
)

func TestRuleFilterAppliesHardFiltersAndScoresSignals(t *testing.T) {
	now := time.Now()
	filter := NewRuleFilter(&model.Config{IncludeKeywords: model.StringSlice{"Go", "backend"}, ExcludeKeywords: model.StringSlice{"manager"}, LocationKeywords: model.StringSlice{"remote"}, WorkTypes: model.StringSlice{"full-time"}})
	match := filter.Evaluate(&model.Role{Title: "Go backend full-time engineer", Location: "Remote", PostedAt: &now})
	if match == nil || match.IncludeScore != 4 || match.BonusScore != 1 || match.Percent == 0 {
		t.Fatalf("match = %#v", match)
	}
	if filter.Evaluate(&model.Role{Title: "Engineering manager", Location: "Remote"}) != nil {
		t.Fatal("excluded role was retained")
	}
	if filter.Evaluate(&model.Role{Title: "Go engineer", Location: "Tokyo"}) != nil {
		t.Fatal("location mismatch was retained")
	}
	if filter.Evaluate(&model.Role{Title: "Go contract engineer", Location: "Remote"}) != nil {
		t.Fatal("non-selected work type was retained")
	}
}

func TestTimeAgoAndKeywordBoundaries(t *testing.T) {
	now := time.Now()
	if TimeAgo(nil) != "" || TimeAgo(&now) != "just now" {
		t.Fatal("unexpected current time rendering")
	}
	for _, tc := range []struct {
		age  time.Duration
		want string
	}{{time.Minute, "1m ago"}, {2 * time.Minute, "2m ago"}, {time.Hour, "1h ago"}, {2 * time.Hour, "2h ago"}, {24 * time.Hour, "1d ago"}, {31 * 24 * time.Hour, "4w ago"}, {400 * 24 * time.Hour, "1y ago"}} {
		then := now.Add(-tc.age)
		if got := TimeAgo(&then); got != tc.want {
			t.Errorf("TimeAgo(%s) = %q, want %q", tc.age, got, tc.want)
		}
	}
	if got := cleanKeywords([]string{" Go ", "", "  ", "Rust"}); len(got) != 2 || got[0] != "Go" || len(cleanKeywords(nil)) != 0 {
		t.Fatalf("clean keywords = %#v", got)
	}
	if newKeywordMatcher("go").Match("gopher") {
		t.Fatal("latin keyword matched inside a word")
	}
	if !newKeywordMatcher("東京").Match("東京都") {
		t.Fatal("CJK keyword did not use substring matching")
	}
}
