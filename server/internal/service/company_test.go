package service

import (
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/testutil"
)

type adapterFake map[string]bool

func (a adapterFake) SupportsAdapter(adapter string) bool { return a[adapter] }

func TestCompanyServiceAggregatesBoardSummary(t *testing.T) {
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
	if err := db.Create(&[]model.CareerBoard{
		{CompanyID: companies[0].ID, Provider: "fake", BoardIdentifier: "summary-fresh-a", CanonicalURL: "https://fresh.test/a", Active: true, LastScanAttemptAt: &now, LastNewRoleDiscoveryAt: &recent},
		{CompanyID: companies[0].ID, Provider: "fake", BoardIdentifier: "summary-fresh-b", CanonicalURL: "https://fresh.test/b", Active: true, LastScanAttemptAt: &recent, LastNewRoleDiscoveryAt: &recent},
		{CompanyID: companies[1].ID, Provider: "fake", BoardIdentifier: "summary-stale", CanonicalURL: "https://stale.test", Active: true, LastScanAttemptAt: &now, LastNewRoleDiscoveryAt: &old},
		{CompanyID: companies[2].ID, Provider: "fake", BoardIdentifier: "summary-none", CanonicalURL: "https://none.test", Active: true},
		{CompanyID: companies[3].ID, Provider: "fake", BoardIdentifier: "summary-failing", CanonicalURL: "https://failing.test", Active: true, LastScanAttemptAt: &now, LastScanFailureDetail: &failure},
		{CompanyID: companies[4].ID, Provider: "fake", BoardIdentifier: "summary-disabled", CanonicalURL: "https://disabled.test", Active: false, LastNewRoleDiscoveryAt: &old},
		{CompanyID: companies[5].ID, Provider: "other", BoardIdentifier: "summary-unsupported", CanonicalURL: "https://unsupported.test", Active: true, LastNewRoleDiscoveryAt: &recent},
	}).Error; err != nil {
		t.Fatalf("create boards: %v", err)
	}
	if err := db.Model(&model.CareerBoard{}).Where("board_identifier = ?", "summary-disabled").Update("active", false).Error; err != nil {
		t.Fatalf("disable board: %v", err)
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
	for name, want := range map[string]string{
		"Fresh": "fresh", "Stale": "stale", "No activity": "no_activity_yet", "Failing": "failing", "Disabled": "not_applicable", "Unsupported": "not_applicable",
	} {
		if item := got[name]; item.FreshnessStatus != want {
			t.Errorf("%s freshness = %s, want %s", name, item.FreshnessStatus, want)
		}
	}
	if got["Fresh"].RoleCount != 2 || got["Fresh"].BoardCount != 2 || got["Stale"].RoleCount != 0 {
		t.Fatalf("summary counts = fresh roles:%d boards:%d stale roles:%d", got["Fresh"].RoleCount, got["Fresh"].BoardCount, got["Stale"].RoleCount)
	}
	if got["Fresh"].LastScanAttemptAt == nil || !got["Fresh"].LastScanAttemptAt.Equal(now) {
		t.Fatalf("latest scan = %v, want %v", got["Fresh"].LastScanAttemptAt, now)
	}
}

func TestCompanyServiceSuggestsExistingBoardOwner(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := model.Company{Name: "PayPay", CareersURL: "https://paypay.example/careers", Active: true}
	if err := db.Create(&company).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&model.CareerBoard{CompanyID: company.ID, Provider: "greenhouse", BoardIdentifier: "paypay", CanonicalURL: "https://job-boards.greenhouse.io/paypay"}).Error; err != nil {
		t.Fatal(err)
	}
	svc := NewCompanyService(repos.Company, adapterFake{})
	name, err := svc.SuggestedEmployerName([]model.CareerBoardDiscoveryCandidate{{Provider: "greenhouse", BoardIdentifier: "paypay"}})
	if err != nil || name != "PayPay" {
		t.Fatalf("SuggestedEmployerName() = %q, %v", name, err)
	}
}
