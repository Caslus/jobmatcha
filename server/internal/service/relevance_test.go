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
	if newKeywordMatcher("go").Match("gopher") {
		t.Fatal("latin keyword matched inside a word")
	}
	if !newKeywordMatcher("東京").Match("東京都") {
		t.Fatal("CJK keyword did not use substring matching")
	}
}
