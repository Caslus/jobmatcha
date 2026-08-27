package repository_test

import (
	"context"
	"testing"

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
