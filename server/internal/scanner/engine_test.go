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
func (p fakeProvider) Fetch(context.Context, *model.Company, *model.CareerBoard) ([]*model.Role, error) {
	return p.roles, p.err
}

func TestEngineScansWithInjectedProviderAndDeduplicatesRoles(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Acme", CareersURL: "https://acme.test", ATSType: "fake", Active: true}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	if err := db.Create(&model.CareerBoard{CompanyID: company.ID, Provider: "fake", BoardIdentifier: "acme", CanonicalURL: company.CareersURL, Active: true}).Error; err != nil {
		t.Fatalf("board: %v", err)
	}
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{roles: []*model.Role{{URL: "https://acme.test/jobs/1", Title: "Go engineer"}}})
	first := engine.ScanAll(context.Background())
	second := engine.ScanAll(context.Background())
	if len(first) != 1 || first[0].NewRoles != 1 || second[0].NewRoles != 0 {
		t.Fatalf("scan results = %#v / %#v", first, second)
	}
	boards, err := repos.CareerBoard.ListForCompany(company.ID)
	if err != nil || len(boards) != 1 || boards[0].LastScanAttemptAt == nil || boards[0].LastSuccessfulScanAt == nil || boards[0].LastNewRoleDiscoveryAt == nil || boards[0].LastScanFailureDetail != nil {
		t.Fatalf("scan evidence = %#v, %v", boards, err)
	}
	count, err := repos.Role.CountAll()
	if err != nil || count != 1 {
		t.Fatalf("role count = %d, %v", count, err)
	}
}

func TestEngineScansMultipleEnabledBoardsForOneCompany(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Multi", CareersURL: "https://multi.test", Active: true}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	for _, slug := range []string{"one", "two"} {
		if err := db.Create(&model.CareerBoard{CompanyID: company.ID, Provider: "fake", BoardIdentifier: slug, CanonicalURL: "https://multi.test/" + slug, Active: true}).Error; err != nil {
			t.Fatalf("board: %v", err)
		}
	}
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{})
	if got := engine.ScanAll(context.Background()); len(got) != 2 {
		t.Fatalf("scans = %#v", got)
	}
}

func TestScanOnlyProviderIsNotRecognized(t *testing.T) {
	registry := NewRegistry()
	registry.Register(fakeProvider{})
	if _, ok := registry.Recognize("https://example.test/board"); ok {
		t.Fatal("scan-only provider was recognized")
	}
}

func TestEngineReportsProviderFailuresWithoutNetwork(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Broken", CareersURL: "https://broken.test", ATSType: "fake", Active: true}
	_ = db.Create(company).Error
	_ = db.Create(&model.CareerBoard{CompanyID: company.ID, Provider: "fake", BoardIdentifier: "broken", CanonicalURL: company.CareersURL, Active: true}).Error
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{err: errors.New("offline")})
	got := engine.ScanAll(context.Background())
	if len(got) != 1 || got[0].Error != "offline" {
		t.Fatalf("results = %#v", got)
	}
	boards, err := repos.CareerBoard.ListForCompany(company.ID)
	if err != nil || len(boards) != 1 || boards[0].LastScanAttemptAt == nil || boards[0].LastSuccessfulScanAt != nil || boards[0].LastScanFailureDetail == nil || *boards[0].LastScanFailureDetail != "offline" {
		t.Fatalf("failure evidence = %#v, %v", boards, err)
	}
}

func TestEngineRecordsSuccessfulEmptyScan(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	company := &model.Company{Name: "Empty", CareersURL: "https://empty.test", ATSType: "fake", Active: true}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("company: %v", err)
	}
	if err := db.Create(&model.CareerBoard{CompanyID: company.ID, Provider: "fake", BoardIdentifier: "empty", CanonicalURL: company.CareersURL, Active: true}).Error; err != nil {
		t.Fatalf("board: %v", err)
	}
	engine := NewEngine(db, repos)
	engine.Registry = NewRegistry()
	engine.Registry.Register(fakeProvider{})
	got := engine.ScanAll(context.Background())
	if len(got) != 1 || got[0].Error != "" || got[0].TotalRoles != 0 {
		t.Fatalf("results = %#v", got)
	}
	boards, err := repos.CareerBoard.ListForCompany(company.ID)
	if err != nil || len(boards) != 1 || boards[0].LastScanAttemptAt == nil || boards[0].LastSuccessfulScanAt == nil || boards[0].LastNewRoleDiscoveryAt != nil || boards[0].LastScanFailureDetail != nil {
		t.Fatalf("empty scan evidence = %#v, %v", boards, err)
	}
}
