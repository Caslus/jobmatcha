package scanner

import (
	"context"
	"errors"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/testutil"
)

type fakeProvider struct {
	roles []*model.Role
	err   error
}

func (fakeProvider) Name() string { return "fake" }
func (p fakeProvider) Fetch(context.Context, *model.Company) ([]*model.Role, error) {
	return p.roles, p.err
}

func TestEngineScansWithInjectedProviderAndDeduplicatesRoles(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Acme", CareersURL: "https://acme.test", ATSType: "fake", Active: true}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{roles: []*model.Role{{URL: "https://acme.test/jobs/1", Title: "Go engineer"}}})
	first := engine.ScanAll(context.Background())
	second := engine.ScanAll(context.Background())
	if len(first) != 1 || first[0].NewRoles != 1 || second[0].NewRoles != 0 {
		t.Fatalf("scan results = %#v / %#v", first, second)
	}
	stored, err := repos.Company.GetByID(company.ID)
	if err != nil || stored.LastScanAttemptAt == nil || stored.LastSuccessfulScanAt == nil || stored.LastNewRoleDiscoveryAt == nil || stored.LastScanFailureDetail != nil {
		t.Fatalf("scan evidence = %#v, %v", stored, err)
	}
	count, err := repos.Role.CountAll()
	if err != nil || count != 1 {
		t.Fatalf("role count = %d, %v", count, err)
	}
}

func TestEngineReportsProviderFailuresWithoutNetwork(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Broken", CareersURL: "https://broken.test", ATSType: "fake", Active: true}
	_ = db.Create(company).Error
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{err: errors.New("offline")})
	got := engine.ScanAll(context.Background())
	if len(got) != 1 || got[0].Error != "offline" {
		t.Fatalf("results = %#v", got)
	}
	stored, err := repos.Company.GetByID(company.ID)
	if err != nil || stored.LastScanAttemptAt == nil || stored.LastSuccessfulScanAt != nil || stored.LastScanFailureDetail == nil || *stored.LastScanFailureDetail != "offline" {
		t.Fatalf("failure evidence = %#v, %v", stored, err)
	}
}

func TestEngineRecordsSuccessfulEmptyScan(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Empty", CareersURL: "https://empty.test", ATSType: "fake", Active: true}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{})
	got := engine.ScanAll(context.Background())
	if len(got) != 1 || got[0].Error != "" || got[0].TotalRoles != 0 {
		t.Fatalf("results = %#v", got)
	}
	stored, err := repos.Company.GetByID(company.ID)
	if err != nil || stored.LastScanAttemptAt == nil || stored.LastSuccessfulScanAt == nil || stored.LastNewRoleDiscoveryAt != nil || stored.LastScanFailureDetail != nil {
		t.Fatalf("empty scan evidence = %#v, %v", stored, err)
	}
}
