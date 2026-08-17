package main

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/caslus/jobmatcha/internal/api"
	"github.com/caslus/jobmatcha/internal/repository"
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
	sqlDB.SetMaxOpenConns(1)

	if err := migrations.Migrate(db); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	if err := migrations.SeedCompanies(db); err != nil {
		slog.Error("failed to seed companies", "error", err)
		os.Exit(1)
	}

	repos := repository.NewRepositories(db)

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	api.RegisterRoutes(r, repos)

	slog.Info("listening", "port", port, "db", dbPath)
	if err := r.Run(":" + port); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}