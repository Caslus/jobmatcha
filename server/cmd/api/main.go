package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/caslus/jobmatcha/internal/api"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/migrations"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
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

	// Ensure config has default admin password (backfill for existing DBs)
	{
		var count int64
		db.Model(&model.Config{}).Count(&count)
		if count == 0 {
			hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), 12)
			db.Create(&model.Config{PasswordHash: string(hash)})
			slog.Info("created default admin password")
		} else {
			// Backfill for existing configs that might have an empty hash
			var cfg model.Config
			db.Where("id = 1").Find(&cfg)
			if cfg.PasswordHash == "" {
				hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), 12)
				db.Model(&model.Config{}).Where("id = 1").Updates(map[string]interface{}{
					"password_hash":  string(hash),
					"setup_complete": false,
				})
				slog.Info("backfilled default admin password")
			}
		}
	}

	repos := repository.NewRepositories(db)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.New(cors.Config{
		AllowOriginFunc: func(origin string) bool {
			// Allow all localhost origins regardless of port (dev servers float)
			return strings.HasPrefix(origin, "http://localhost:")
		},
		AllowMethods:     []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Cookie"},
		AllowCredentials: true,
	}))

	schedulerSvc := api.RegisterRoutes(r, repos, db)

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