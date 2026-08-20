package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/go-co-op/gocron/v2"
)

// SchedulerService manages a gocron scheduler that triggers scans on a cron schedule.
// The schedule and timezone are stored in the Config model and can be changed at runtime
// via ReloadSchedule.
type SchedulerService struct {
	mu        sync.Mutex
	scheduler gocron.Scheduler
	cfgRepo   *repository.ConfigRepo
	scanner   *ScannerService
	jobID     gocron.Job
	started   bool
	tzString  string
}

// NewSchedulerService creates a scheduler service backed by gocron.
// The scheduler is not started; call Start() to begin.
func NewSchedulerService(cfgRepo *repository.ConfigRepo, scanner *ScannerService) *SchedulerService {
	return &SchedulerService{
		cfgRepo: cfgRepo,
		scanner: scanner,
	}
}

// Start reads the current config and starts the scheduler loop.
// If scans are enabled, it registers a cron job and starts ticking.
func (s *SchedulerService) Start() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cfg, err := s.cfgRepo.Get()
	if err != nil {
		slog.Error("scheduler: failed to read config on start", "error", err)
		return
	}

	s.createSchedulerLocked(cfg.ScanTimezone)

	if cfg.ScanEnabled {
		s.registerJobLocked(cfg.ScanCronExpr)
	}

	s.scheduler.Start()
	s.started = true
	slog.Info("scheduler started",
		"scan_enabled", cfg.ScanEnabled,
		"cron_expr", cfg.ScanCronExpr,
		"timezone", cfg.ScanTimezone,
	)
}

// Stop gracefully shuts down the scheduler, waiting for any in-flight scan to finish.
func (s *SchedulerService) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	if err := s.scheduler.Shutdown(); err != nil {
		slog.Error("scheduler: shutdown error", "error", err)
	}
	s.started = false
	s.jobID = nil
	slog.Info("scheduler stopped")
}

// ReloadSchedule re-reads the config and adjusts the job accordingly.
// Call this after PUT /api/settings changes any scan config fields.
func (s *SchedulerService) ReloadSchedule() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.started {
		return
	}

	// Shutdown the entire scheduler (waits for any in-flight job)
	if err := s.scheduler.Shutdown(); err != nil {
		slog.Error("scheduler: shutdown during reload", "error", err)
	}
	s.jobID = nil

	cfg, err := s.cfgRepo.Get()
	if err != nil {
		slog.Error("scheduler: failed to read config on reload", "error", err)
		return
	}

	// Recreate with (possibly new) timezone
	s.createSchedulerLocked(cfg.ScanTimezone)

	if cfg.ScanEnabled {
		s.registerJobLocked(cfg.ScanCronExpr)
	}

	s.scheduler.Start()
	slog.Info("scheduler reloaded",
		"scan_enabled", cfg.ScanEnabled,
		"cron_expr", cfg.ScanCronExpr,
		"timezone", cfg.ScanTimezone,
	)
}

// IsEnabled returns whether scheduled scanning is currently active.
func (s *SchedulerService) IsEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.jobID != nil
}

// NextRun returns the next scheduled scan time, or nil if no job is registered.
func (s *SchedulerService) NextRun() *time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.jobID == nil {
		return nil
	}
	t, err := s.jobID.NextRun()
	if err != nil {
		return nil
	}
	return &t
}

// createSchedulerLocked sets up a new gocron scheduler for the given timezone.
// Caller must hold s.mu.
func (s *SchedulerService) createSchedulerLocked(tz string) {
	if s.jobID != nil {
		s.jobID = nil
	}

	loc := parseLocation(tz)
	s.tzString = tz

	sched, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		slog.Error("scheduler: failed to create scheduler", "timezone", tz, "error", err)
		// Fall back to UTC
		sched, _ = gocron.NewScheduler(gocron.WithLocation(time.UTC))
		s.tzString = "UTC"
	}
	s.scheduler = sched
}

// registerJobLocked creates a new gocron job. Caller must hold s.mu.
func (s *SchedulerService) registerJobLocked(cronExpr string) {
	job, err := s.scheduler.NewJob(
		gocron.CronJob(cronExpr, false), // 5-field standard cron (no seconds)
		gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
			defer cancel()

			slog.Info("scheduler: firing scan", "cron", cronExpr, "tz", s.tzString)
			jobID, err := s.scanner.StartScan()
			if err != nil {
				slog.Error("scheduler: scan failed to start", "error", err)
				return
			}
			_ = ctx // keep for future use (e.g. passing context to StartScan)
			slog.Info("scheduler: scan started", "job_id", jobID)
		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
		gocron.WithName("scan"),
	)
	if err != nil {
		slog.Error("scheduler: failed to register job", "cron", cronExpr, "error", err)
		return
	}
	s.jobID = job
}

// parseLocation converts a timezone name (e.g. "Asia/Tokyo") to *time.Location.
// Falls back to UTC on error.
func parseLocation(tz string) *time.Location {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		slog.Warn("scheduler: invalid timezone, falling back to UTC", "timezone", tz, "error", err)
		return time.UTC
	}
	return loc
}