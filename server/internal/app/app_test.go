package app

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/ai"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"gorm.io/gorm"
)

type fakeScanner struct{}

func (fakeScanner) StartScan() (uint, error)                      { return 0, nil }
func (fakeScanner) GetJob(uint) (*model.ScanJobResponse, error)   { return nil, nil }
func (fakeScanner) GetLatestJob() (*model.ScanJobResponse, error) { return nil, nil }

type fakeScheduler struct{ starts, stops int }

func (s *fakeScheduler) Start()              { s.starts++ }
func (s *fakeScheduler) Stop()               { s.stops++ }
func (s *fakeScheduler) ReloadSchedule()     {}
func (s *fakeScheduler) IsEnabled() bool     { return false }
func (s *fakeScheduler) NextRun() *time.Time { return nil }

type fakeAI struct{}

func (fakeAI) ValidateKey(context.Context, string) (bool, int, error) { return true, 1, nil }
func (fakeAI) ParseResume(context.Context, string, string) (*ai.ParseResumeResult, error) {
	return nil, nil
}
func (fakeAI) TailorResume(context.Context, string, model.ResumeDocument, string, string, string, string) (*model.ResumeDocument, error) {
	return nil, nil
}

func TestNewUsesInjectableSeamsAndCleanlyCloses(t *testing.T) {
	scheduler := &fakeScheduler{}
	application, err := New(context.Background(), Options{
		DBPath:           filepath.Join(t.TempDir(), "app.db"),
		DisableScheduler: true,
		AI:               fakeAI{},
		NewScanner: func(*gorm.DB, *repository.Repositories) service.Scanner {
			return fakeScanner{}
		},
		NewScheduler: func(*repository.ConfigRepo, service.Scanner) service.Scheduler {
			return scheduler
		},
	})
	if err != nil {
		t.Fatalf("new application: %v", err)
	}
	if scheduler.starts != 0 {
		t.Fatalf("scheduler starts = %d, want 0", scheduler.starts)
	}
	w := httptest.NewRecorder()
	application.Router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("health status = %d, want %d", w.Code, http.StatusOK)
	}
	if err := application.Close(); err != nil {
		t.Fatalf("close application: %v", err)
	}
	if scheduler.stops != 1 {
		t.Fatalf("scheduler stops = %d, want 1", scheduler.stops)
	}
}

func TestNewStartsSchedulerByDefault(t *testing.T) {
	scheduler := &fakeScheduler{}
	application, err := New(context.Background(), Options{
		DBPath:       filepath.Join(t.TempDir(), "app.db"),
		NewScanner:   func(*gorm.DB, *repository.Repositories) service.Scanner { return fakeScanner{} },
		NewScheduler: func(*repository.ConfigRepo, service.Scanner) service.Scheduler { return scheduler },
	})
	if err != nil {
		t.Fatalf("new application: %v", err)
	}
	defer application.Close()
	if scheduler.starts != 1 {
		t.Fatalf("scheduler starts = %d, want 1", scheduler.starts)
	}
}

func TestNewRejectsInvalidStartupConfiguration(t *testing.T) {
	t.Run("requires a database path", func(t *testing.T) {
		_, err := New(context.Background(), Options{})
		if err == nil || err.Error() != "application database path is required" {
			t.Fatalf("New error = %v", err)
		}
	})

	t.Run("reports migration failure", func(t *testing.T) {
		migrateErr := errors.New("migration failed")
		_, err := New(context.Background(), Options{
			DBPath:  filepath.Join(t.TempDir(), "app.db"),
			Migrate: func(*gorm.DB) error { return migrateErr },
		})
		if !errors.Is(err, migrateErr) {
			t.Fatalf("New error = %v, want migration error", err)
		}
	})

	t.Run("reports seed failure", func(t *testing.T) {
		seedErr := errors.New("seed failed")
		_, err := New(context.Background(), Options{
			DBPath:  filepath.Join(t.TempDir(), "app.db"),
			Migrate: func(*gorm.DB) error { return nil },
			SeedCompanies: func(*gorm.DB) error {
				return seedErr
			},
		})
		if !errors.Is(err, seedErr) {
			t.Fatalf("New error = %v, want seed error", err)
		}
	})

	t.Run("reports missing static frontend", func(t *testing.T) {
		_, err := New(context.Background(), Options{
			DBPath:           filepath.Join(t.TempDir(), "app.db"),
			StaticDir:        t.TempDir(),
			DisableScheduler: true,
			NewScanner:       func(*gorm.DB, *repository.Repositories) service.Scanner { return fakeScanner{} },
			NewScheduler:     func(*repository.ConfigRepo, service.Scanner) service.Scheduler { return &fakeScheduler{} },
		})
		if err == nil {
			t.Fatal("New error = nil, want missing static frontend error")
		}
	})
}
