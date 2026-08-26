// Package app composes Jobmatcha's production dependencies into a runnable HTTP application.
package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/caslus/jobmatcha/internal/api"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/caslus/jobmatcha/migrations"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type ScannerFactory func(*gorm.DB, *repository.Repositories) service.Scanner
type SchedulerFactory func(*repository.ConfigRepo, service.Scanner) service.Scheduler

// Options exposes only process-boundary collaborators. Nil factories retain
// production behavior; tests can replace the AI, ATS, and scheduler seams.
type Options struct {
	DBPath                string
	BootstrapPasswordFile string
	StaticDir             string
	CookieSecure          bool
	DisableScheduler      bool
	AI                    service.AIClient
	NewScanner            ScannerFactory
	NewScheduler          SchedulerFactory
	Migrate               func(*gorm.DB) error
	SeedCompanies         func(*gorm.DB) error
}

func ensureDir(path string) error {
	return os.MkdirAll(path, 0755)
}

// Application owns the router and resources that must be stopped at shutdown.
type Application struct {
	Router    *gin.Engine
	DB        *gorm.DB
	Repos     *repository.Repositories
	Scheduler service.Scheduler
	sqlDB     interface{ Close() error }
	stopOnce  sync.Once
}

func New(ctx context.Context, options Options) (*Application, error) {
	if options.DBPath == "" {
		return nil, fmt.Errorf("application database path is required")
	}
	if options.BootstrapPasswordFile == "" {
		options.BootstrapPasswordFile = service.BootstrapPasswordFile(options.DBPath)
	}
	if options.AI == nil {
		options.AI = service.NewAIClient()
	}
	if options.NewScanner == nil {
		options.NewScanner = func(db *gorm.DB, repos *repository.Repositories) service.Scanner {
			return service.NewScannerService(db, repos)
		}
	}
	if options.NewScheduler == nil {
		options.NewScheduler = func(config *repository.ConfigRepo, scanner service.Scanner) service.Scheduler {
			return service.NewSchedulerService(config, scanner)
		}
	}
	if options.Migrate == nil {
		options.Migrate = migrations.Migrate
	}
	if options.SeedCompanies == nil {
		options.SeedCompanies = migrations.SeedCompanies
	}

	dir := filepath.Dir(options.DBPath)
	if dir != "." {
		if err := ensureDir(dir); err != nil {
			return nil, fmt.Errorf("creating database directory: %w", err)
		}
	}
	db, err := gorm.Open(sqlite.Open(options.DBPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("getting database handle: %w", err)
	}
	sqlDB.SetMaxOpenConns(3)
	if err := options.Migrate(db); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("migrating database: %w", err)
	}
	if err := options.SeedCompanies(db); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("seeding companies: %w", err)
	}

	repos := repository.NewRepositories(db)
	auth := service.NewAuthService(repos, options.BootstrapPasswordFile)
	if err := auth.EnsureBootstrap(ctx); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("initializing authentication: %w", err)
	}
	scanner := options.NewScanner(db, repos)
	scheduler := options.NewScheduler(repos.Config, scanner)

	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	api.RegisterRoutes(router, repos, api.RouteDependencies{
		Auth: auth, Scanner: scanner, Scheduler: scheduler, AI: options.AI, CookieSecure: options.CookieSecure,
	})
	if options.StaticDir != "" {
		if err := api.RegisterStaticRoutes(router, options.StaticDir); err != nil {
			sqlDB.Close()
			return nil, fmt.Errorf("registering static frontend: %w", err)
		}
	}
	if !options.DisableScheduler {
		scheduler.Start()
	}

	return &Application{Router: router, DB: db, Repos: repos, Scheduler: scheduler, sqlDB: sqlDB}, nil
}

// Stop prevents new scheduled work. It is safe to call more than once.
func (a *Application) Stop() {
	a.stopOnce.Do(func() {
		if a.Scheduler != nil {
			a.Scheduler.Stop()
		}
	})
}

// Close stops future scheduled work before releasing database resources.
func (a *Application) Close() error {
	a.Stop()
	if a.sqlDB != nil {
		return a.sqlDB.Close()
	}
	return nil
}
