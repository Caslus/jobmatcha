package service

import (
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/testutil"
)

type adapterFake map[string]bool

func (a adapterFake) SupportsAdapter(adapter string) bool { return a[adapter] }

func TestCompanyServiceClassifiesAdapterAndFreshness(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	now := time.Date(2026, 8, 31, 12, 0, 0, 0, time.UTC)
	old := now.Add(-31 * 24 * time.Hour)
	recent := now.Add(-24 * time.Hour)
	failure := "request failed"
	companies := []model.Company{
		{Name: "Fresh", CareersURL: "https://fresh.test", ATSType: "fake", Active: true, LastScanAttemptAt: &now, LastNewRoleDiscoveryAt: &recent},
		{Name: "Stale", CareersURL: "https://stale.test", ATSType: "fake", Active: true, LastScanAttemptAt: &now, LastNewRoleDiscoveryAt: &old},
		{Name: "No activity", CareersURL: "https://none.test", ATSType: "fake", Active: true},
		{Name: "Failing", CareersURL: "https://failing.test", ATSType: "fake", Active: true, LastScanAttemptAt: &now, LastScanFailureDetail: &failure},
		{Name: "Disabled", CareersURL: "https://disabled.test", ATSType: "fake", Active: false, LastNewRoleDiscoveryAt: &old},
		{Name: "Unsupported", CareersURL: "https://unsupported.test", ATSType: "other", Active: true, LastNewRoleDiscoveryAt: &recent},
	}
	if err := db.Create(&companies).Error; err != nil {
		t.Fatalf("create companies: %v", err)
	}
	if err := db.Model(&model.Company{}).Where("name = ?", "Disabled").Update("active", false).Error; err != nil {
		t.Fatalf("disable company: %v", err)
	}
	if err := db.Create(&[]model.Role{
		{CompanyID: companies[0].ID, URLHash: "fresh-1", URL: "https://fresh.test/1", Title: "Fresh role"},
		{CompanyID: companies[0].ID, URLHash: "fresh-2", URL: "https://fresh.test/2", Title: "Another fresh role"},
	}).Error; err != nil {
		t.Fatalf("create roles: %v", err)
	}
	svc := NewCompanyService(repos.Company, adapterFake{"fake": true})
	svc.now = func() time.Time { return now }
	items, err := svc.List()
	if err != nil || len(items) != len(companies) {
		t.Fatalf("list = %#v, %v", items, err)
	}
	got := map[string]model.CompanyListItem{}
	for _, item := range items {
		got[item.Name] = item
	}
	for name, want := range map[string][2]string{
		"Fresh": {"healthy", "fresh"}, "Stale": {"healthy", "stale"}, "No activity": {"unknown", "no_activity_yet"}, "Failing": {"failing", "no_activity_yet"}, "Disabled": {"unknown", "not_applicable"}, "Unsupported": {"unsupported", "not_applicable"},
	} {
		if item := got[name]; item.AdapterStatus != want[0] || item.FreshnessStatus != want[1] {
			t.Errorf("%s = %s/%s, want %s/%s", name, item.AdapterStatus, item.FreshnessStatus, want[0], want[1])
		}
	}
	if got["Fresh"].RoleCount != 2 || got["Stale"].RoleCount != 0 {
		t.Fatalf("role counts = fresh:%d stale:%d", got["Fresh"].RoleCount, got["Stale"].RoleCount)
	}
}
