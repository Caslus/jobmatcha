package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/caslus/jobmatcha/internal/api"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/caslus/jobmatcha/migrations"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/app.db"
	}

	dir := filepath.Dir(dbPath)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			slog.Error("failed to create db directory", "error", err)
			os.Exit(1)
		}
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8181"
	}

	db, err := gorm.Open(sqlite.Open(dbPath+"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"), &gorm.Config{})
	if err != nil {
		slog.Error("failed to open database", "error", err)
		os.Exit(1)
	}

	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get underlying sql.DB", "error", err)
		os.Exit(1)
	}
	sqlDB.SetMaxOpenConns(3)

	if err := migrations.Migrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if err := migrations.SeedCompanies(db); err != nil {
		slog.Error("failed to seed companies", "error", err)
		os.Exit(1)
	}

	repos := repository.NewRepositories(db)
	authSvc := service.NewAuthService(repos, service.BootstrapPasswordFile(dbPath))
	if err := authSvc.EnsureBootstrap(context.Background()); err != nil {
		slog.Error("failed to initialize authentication", "error", err)
		os.Exit(1)
	}

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	schedulerSvc := api.RegisterRoutes(r, repos, db, authSvc)
	if staticDir := findStaticDir(); staticDir != "" {
		if err := api.RegisterStaticRoutes(r, staticDir); err != nil {
			slog.Error("failed to register static frontend", "dir", staticDir, "error", err)
			os.Exit(1)
		}
		slog.Info("serving static frontend", "dir", staticDir)
	} else {
		slog.Warn("static frontend not found; serving API only", "hint", "build web or set STATIC_DIR")
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	// Start server in background
	go func() {
		slog.Info("listening", "port", port, "db", dbPath)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down...")

	// Stop the scheduler first (no new scans)
	schedulerSvc.Stop()

	// Then shut down the HTTP server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server forced to shutdown", "error", err)
	}

	slog.Info("server stopped")
}

func findStaticDir() string {
	if configured := os.Getenv("STATIC_DIR"); configured != "" {
		return configured
	}

	for _, candidate := range []string{"web/dist/client", "../web/dist/client"} {
		if _, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil {
			return candidate
		}
	}

	return ""
}
