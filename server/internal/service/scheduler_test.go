package service

import (
	"context"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/testutil"
)

type schedulerScannerFake struct{ starts int }

func (s *schedulerScannerFake) StartScan() (uint, error)                    { s.starts++; return 1, nil }
func (*schedulerScannerFake) GetJob(uint) (*model.ScanJobResponse, error)   { return nil, nil }
func (*schedulerScannerFake) GetLatestJob() (*model.ScanJobResponse, error) { return nil, nil }
func (*schedulerScannerFake) SupportsAdapter(string) bool                   { return false }

func TestSchedulerRegistersAndRemovesConfiguredJob(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	if err := repos.Config.Create(context.Background(), &model.Config{ID: 1, ScanEnabled: true, ScanCronExpr: "0 * * * *", ScanTimezone: "UTC"}); err != nil {
		t.Fatalf("config: %v", err)
	}
	service := NewSchedulerService(repos.Config, &schedulerScannerFake{})
	service.Start()
	defer service.Stop()
	if !service.IsEnabled() || service.NextRun() == nil {
		t.Fatal("enabled schedule was not registered")
	}
	if err := repos.Config.UpdateMap(context.Background(), map[string]interface{}{"scan_enabled": false}); err != nil {
		t.Fatalf("disable: %v", err)
	}
	service.ReloadSchedule()
	if service.IsEnabled() || service.NextRun() != nil {
		t.Fatal("disabled schedule retained a job")
	}
}

func TestSchedulerDoesNotReloadBeforeStartOrAfterStop(t *testing.T) {
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	ctx := context.Background()
	if err := repos.Config.Create(ctx, &model.Config{ID: 1, ScanEnabled: false, ScanCronExpr: "0 * * * *", ScanTimezone: "UTC"}); err != nil {
		t.Fatalf("config: %v", err)
	}

	service := NewSchedulerService(repos.Config, &schedulerScannerFake{})
	service.ReloadSchedule()
	if service.IsEnabled() || service.NextRun() != nil {
		t.Fatal("reload before start registered a job")
	}

	service.Start()
	if service.IsEnabled() || service.NextRun() != nil {
		t.Fatal("disabled schedule registered a job")
	}
	service.Stop()
	service.Stop() // Stopping an already stopped scheduler must be safe.

	if err := repos.Config.UpdateMap(ctx, map[string]interface{}{"scan_enabled": true}); err != nil {
		t.Fatalf("enable: %v", err)
	}
	service.ReloadSchedule()
	if service.IsEnabled() || service.NextRun() != nil {
		t.Fatal("reload after stop restarted the scheduler")
	}
}

func TestParseLocationFallsBackToUTC(t *testing.T) {
	if got := parseLocation("not/a/timezone"); got.String() != "UTC" {
		t.Fatalf("fallback = %s", got)
	}
	if got := parseLocation("America/Sao_Paulo"); got.String() != "America/Sao_Paulo" {
		t.Fatalf("location = %s", got)
	}
}
