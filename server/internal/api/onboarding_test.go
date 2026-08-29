package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/caslus/jobmatcha/internal/testutil"
	"github.com/gin-gonic/gin"
)

type onboardingScannerFake struct {
	jobID uint
	err   error
	calls int
}

func (f *onboardingScannerFake) StartScan() (uint, error) {
	f.calls++
	return f.jobID, f.err
}

func (f *onboardingScannerFake) GetJob(uint) (*model.ScanJobResponse, error) { return nil, nil }

func (f *onboardingScannerFake) GetLatestJob() (*model.ScanJobResponse, error) { return nil, nil }

var _ service.Scanner = (*onboardingScannerFake)(nil)

type onboardingSchedulerFake struct{ reloads int }

func (f *onboardingSchedulerFake) Start()              {}
func (f *onboardingSchedulerFake) Stop()               {}
func (f *onboardingSchedulerFake) ReloadSchedule()     { f.reloads++ }
func (f *onboardingSchedulerFake) IsEnabled() bool     { return false }
func (f *onboardingSchedulerFake) NextRun() *time.Time { return nil }

var _ service.Scheduler = (*onboardingSchedulerFake)(nil)

func setupOnboardingHandler(t *testing.T, scanner *onboardingScannerFake, scheduler *onboardingSchedulerFake) (*OnboardingHandler, *repository.ConfigRepo) {
	t.Helper()
	db := testutil.Database(t)
	repo := repository.NewConfigRepo(db)
	if err := repo.Create(context.Background(), &model.Config{ID: 1}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	return NewOnboardingHandler(repo, scanner, scheduler), repo
}

func performOnboarding(t *testing.T, handler *OnboardingHandler, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/onboarding/complete", handler.Complete)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/onboarding/complete", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(recorder, request)
	return recorder
}

func TestOnboardingHandlerComplete(t *testing.T) {
	scanner := &onboardingScannerFake{jobID: 42}
	scheduler := &onboardingSchedulerFake{}
	handler, repo := setupOnboardingHandler(t, scanner, scheduler)

	response := performOnboarding(t, handler, `{
		"user_name":"Ada Lovelace",
		"user_email":"ada@example.com",
		"user_location":"London",
		"user_linkedin":"https://linkedin.example/ada",
		"user_github":"https://github.example/ada",
		"include_keywords":["go","backend"],
		"exclude_keywords":["sales"],
		"location_keywords":["remote"],
		"work_types":["full-time"],
		"max_days_old":14,
		"scan_enabled":true,
		"scan_cron_expr":"0 */4 * * *",
		"scan_timezone":"Europe/London"
	}`)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		Status        string `json:"status"`
		SetupComplete bool   `json:"setup_complete"`
		ScanStarted   bool   `json:"scan_started"`
		ScanID        uint   `json:"scan_id"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Status != "ok" || !payload.SetupComplete || !payload.ScanStarted || payload.ScanID != 42 {
		t.Errorf("unexpected response: %+v", payload)
	}
	if scanner.calls != 1 || scheduler.reloads != 1 {
		t.Errorf("scanner calls = %d, scheduler reloads = %d; want 1 each", scanner.calls, scheduler.reloads)
	}

	config, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("get saved config: %v", err)
	}
	if !config.SetupComplete || config.UserName != "Ada Lovelace" || config.UserEmail != "ada@example.com" || config.MaxDaysOld != 14 {
		t.Errorf("unexpected saved config: %+v", config)
	}
	if !config.ScanEnabled || config.ScanCronExpr != "0 */4 * * *" || config.ScanTimezone != "Europe/London" {
		t.Errorf("scan settings were not saved: %+v", config)
	}
	if got := []string(config.IncludeKeywords); len(got) != 2 || got[0] != "go" || got[1] != "backend" {
		t.Errorf("include keywords = %v", got)
	}
}

func TestOnboardingHandlerRejectsInvalidJSON(t *testing.T) {
	scanner := &onboardingScannerFake{}
	scheduler := &onboardingSchedulerFake{}
	handler, repo := setupOnboardingHandler(t, scanner, scheduler)

	response := performOnboarding(t, handler, `{not-json`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	var payload model.ErrorResponse
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Error != "Invalid request." {
		t.Errorf("error = %q", payload.Error)
	}
	if scanner.calls != 0 || scheduler.reloads != 0 {
		t.Errorf("invalid JSON started work: scanner calls = %d, scheduler reloads = %d", scanner.calls, scheduler.reloads)
	}
	config, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("get config: %v", err)
	}
	if config.SetupComplete {
		t.Error("invalid request completed onboarding")
	}
}

func TestOnboardingHandlerKeepsSavedDataWhenFirstScanFails(t *testing.T) {
	scanner := &onboardingScannerFake{err: errors.New("scanner unavailable")}
	scheduler := &onboardingSchedulerFake{}
	handler, repo := setupOnboardingHandler(t, scanner, scheduler)

	response := performOnboarding(t, handler, `{"user_name":"Ada","user_email":"ada@example.com","scan_enabled":false}`)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d: %s", response.Code, response.Body.String())
	}
	var payload struct {
		SetupComplete bool   `json:"setup_complete"`
		ScanStarted   bool   `json:"scan_started"`
		Message       string `json:"message"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !payload.SetupComplete || payload.ScanStarted || payload.Message == "" {
		t.Errorf("unexpected response: %+v", payload)
	}
	if scanner.calls != 1 || scheduler.reloads != 1 {
		t.Errorf("scanner calls = %d, scheduler reloads = %d; want 1 each", scanner.calls, scheduler.reloads)
	}
	config, err := repo.Get(context.Background())
	if err != nil {
		t.Fatalf("get saved config: %v", err)
	}
	if !config.SetupComplete || config.UserName != "Ada" || config.UserEmail != "ada@example.com" {
		t.Errorf("saved config = %+v", config)
	}
}
