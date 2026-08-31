package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/testutil"
)

func TestConfigAndScanJobRepositoriesPersistLifecycle(t *testing.T) {
	db := testutil.Database(t)
	repos := repository.NewRepositories(db)
	if _, err := repos.Config.Get(context.Background()); err != repository.ErrNotFound {
		t.Fatalf("missing config error = %v", err)
	}
	if err := repos.Config.Create(context.Background(), &model.Config{ID: 1}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	cfg, err := repos.Config.Get(context.Background())
	if err != nil || cfg.ScanCronExpr == "" || cfg.ScanTimezone != "UTC" || cfg.IncludeKeywords == nil {
		t.Fatalf("normalized config = %#v, %v", cfg, err)
	}

	job := &model.ScanJob{Status: "pending"}
	if err := repos.ScanJob.Create(job); err != nil {
		t.Fatalf("create job: %v", err)
	}
	if err := repos.ScanJob.UpdateProgress(job.ID, 1, 2); err != nil {
		t.Fatalf("progress: %v", err)
	}
	if err := repos.ScanJob.Complete(job.ID, `[{"company_name":"Acme","new_roles":1,"total_roles":2}]`, 3); err != nil {
		t.Fatalf("complete: %v", err)
	}
	stored, err := repos.ScanJob.GetLatest()
	if err != nil || stored.Status != "completed" || stored.CompletedCompanies != 2 || stored.CompletedAt == nil {
		t.Fatalf("stored job = %#v, %v", stored, err)
	}
	if err := repos.ScanJob.Fail(job.ID, "retry failed"); err != nil {
		t.Fatalf("fail: %v", err)
	}
	stored, _ = repos.ScanJob.GetByID(job.ID)
	if stored.Status != "failed" || stored.Error != "retry failed" {
		t.Fatalf("failed job = %#v", stored)
	}
}

func TestCompanyRepositoryListsActiveCompaniesAndFindsByID(t *testing.T) {
	db := testutil.Database(t)
	repos := repository.NewRepositories(db)

	companies := []model.Company{
		{Name: "Zulu", CareersURL: "https://zulu.example/jobs", Active: true},
		{Name: "Alpha", CareersURL: "https://alpha.example/jobs", Active: false},
		{Name: "Bravo", CareersURL: "https://bravo.example/jobs", Active: true},
	}
	for i := range companies {
		if err := db.Create(&companies[i]).Error; err != nil {
			t.Fatalf("create company %q: %v", companies[i].Name, err)
		}
	}
	if err := db.Model(&model.Company{}).Where("id = ?", companies[1].ID).Update("active", false).Error; err != nil {
		t.Fatalf("deactivate company: %v", err)
	}

	active, err := repos.Company.ListActive()
	if err != nil || len(active) != 2 || active[0].Name != "Bravo" || active[1].Name != "Zulu" {
		t.Fatalf("active companies = %#v, %v", active, err)
	}
	all, err := repos.Company.ListAll()
	if err != nil || len(all) != 3 || all[0].Name != "Alpha" {
		t.Fatalf("all companies = %#v, %v", all, err)
	}

	found, err := repos.Company.GetByID(companies[2].ID)
	if err != nil || found == nil || found.CareersURL != companies[2].CareersURL {
		t.Fatalf("company by ID = %#v, %v", found, err)
	}
	missing, err := repos.Company.GetByID(999)
	if err != nil || missing != nil {
		t.Fatalf("missing company = %#v, %v", missing, err)
	}
}

func TestCompanyRepositoryPersistsScanEvidence(t *testing.T) {
	db := testutil.Database(t)
	repos := repository.NewRepositories(db)
	company := &model.Company{Name: "Evidence", CareersURL: "https://evidence.example/jobs"}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	attemptedAt := time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC)
	discoveredAt := attemptedAt.Add(time.Hour)
	if err := repos.Company.RecordScanAttempt(company.ID, attemptedAt); err != nil {
		t.Fatalf("record attempt: %v", err)
	}
	if err := repos.Company.RecordScanFailure(company.ID, "provider offline"); err != nil {
		t.Fatalf("record failure: %v", err)
	}
	if err := repos.Company.RecordScanSuccess(company.ID, attemptedAt); err != nil {
		t.Fatalf("record success: %v", err)
	}
	if err := repos.Company.RecordNewRoleDiscovery(company.ID, discoveredAt); err != nil {
		t.Fatalf("record discovery: %v", err)
	}
	found, err := repos.Company.GetByID(company.ID)
	if err != nil || found == nil || found.LastScanAttemptAt == nil || !found.LastScanAttemptAt.Equal(attemptedAt) || found.LastSuccessfulScanAt == nil || !found.LastSuccessfulScanAt.Equal(attemptedAt) || found.LastScannedAt == nil || !found.LastScannedAt.Equal(attemptedAt) || found.LastScanFailureDetail != nil || found.LastNewRoleDiscoveryAt == nil || !found.LastNewRoleDiscoveryAt.Equal(discoveredAt) {
		t.Fatalf("stored evidence = %#v, %v", found, err)
	}
}

func TestRoleRepositoryPersistsListsAndPatchesRoles(t *testing.T) {
	db := testutil.Database(t)
	repos := repository.NewRepositories(db)
	company := model.Company{Name: "Acme", CareersURL: "https://acme.example/jobs"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	roles := []model.Role{
		{CompanyID: company.ID, URLHash: "role-one", URL: "https://acme.example/1", Title: "First"},
		{CompanyID: company.ID, URLHash: "role-two", URL: "https://acme.example/2", Title: "Second"},
	}
	if err := repos.Role.BulkCreate(roles); err != nil {
		t.Fatalf("bulk create roles: %v", err)
	}
	if err := repos.Role.BulkCreate(nil); err != nil {
		t.Fatalf("empty bulk create: %v", err)
	}
	third := &model.Role{CompanyID: company.ID, URLHash: "role-three", URL: "https://acme.example/3", Title: "Third"}
	if err := repos.Role.Create(third); err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := db.Model(&model.Role{}).Where("id = ?", roles[0].ID).Update("created_at", time.Now().Add(-time.Hour)).Error; err != nil {
		t.Fatalf("set creation time: %v", err)
	}

	page, total, err := repos.Role.List(1, 2)
	if err != nil || total != 3 || len(page) != 2 || page[0].Company.Name != company.Name {
		t.Fatalf("first page = %#v, total %d, err %v", page, total, err)
	}
	secondPage, total, err := repos.Role.List(2, 2)
	if err != nil || total != 3 || len(secondPage) != 1 {
		t.Fatalf("second page = %#v, total %d, err %v", secondPage, total, err)
	}
	all, err := repos.Role.ListAll()
	if err != nil || len(all) != 3 || all[0].Company.ID != company.ID {
		t.Fatalf("all roles = %#v, %v", all, err)
	}
	if err := repos.Role.Patch(third.ID, map[string]interface{}{"is_hidden": true, "title": "Updated"}); err != nil {
		t.Fatalf("patch role: %v", err)
	}
	found, err := repos.Role.GetByID(third.ID)
	if err != nil || found == nil || !found.IsHidden || found.Title != "Updated" || found.Company.Name != company.Name {
		t.Fatalf("role by ID = %#v, %v", found, err)
	}
	missing, err := repos.Role.GetByID(999)
	if err != nil || missing != nil {
		t.Fatalf("missing role = %#v, %v", missing, err)
	}
	count, err := repos.Role.CountAll()
	if err != nil || count != 3 {
		t.Fatalf("role count = %d, %v", count, err)
	}
}

func TestConfigRepositoryUpdatesPersistedValues(t *testing.T) {
	db := testutil.Database(t)
	repo := repository.NewConfigRepo(db)
	ctx := context.Background()
	if err := repo.Create(ctx, &model.Config{ID: 1}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	updates := map[string]interface{}{
		"ai_enabled":       true,
		"scan_cron_expr":   "0 9 * * 1",
		"include_keywords": model.StringSlice{"go", "react"},
	}
	if err := repo.UpdateMap(ctx, updates); err != nil {
		t.Fatalf("update config: %v", err)
	}
	cfg, err := repo.Get(ctx)
	if err != nil || !cfg.AIEnabled || cfg.ScanCronExpr != "0 9 * * 1" || len(cfg.IncludeKeywords) != 2 {
		t.Fatalf("updated config = %#v, %v", cfg, err)
	}
}
