// Package testutil contains fixtures shared by backend tests.
package testutil

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/migrations"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Database creates a migrated, file-backed SQLite database owned by t.
func Database(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "app.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get test database handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := migrations.Migrate(db); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	return db
}

// Repositories returns the project's real repositories for db.
func Repositories(db *gorm.DB) *repository.Repositories {
	return repository.NewRepositories(db)
}

// Context returns a context suitable for repository fixture setup.
func Context(t *testing.T) context.Context {
	t.Helper()
	return context.Background()
}
