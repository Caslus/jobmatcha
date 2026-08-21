package migrations

import (
	"log/slog"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Company{},
		&model.Role{},
		&model.Config{},
		&model.Session{},
		&model.ScanJob{},
	); err != nil {
		return err
	}

	// Backfill columns that may not exist on older databases
	// (GORM AutoMigrate doesn't always add columns to existing SQLite tables)
	backfillConfigColumns(db)

	return nil
}

func backfillConfigColumns(db *gorm.DB) {
	type colDef struct {
		name string
		typ  string
	}
	cols := []colDef{
		{"ai_provider", "TEXT DEFAULT ''"},
		{"ai_api_key", "TEXT DEFAULT ''"},
		{"ai_enabled", "INTEGER DEFAULT 0"},
		{"user_name", "TEXT DEFAULT ''"},
		{"user_email", "TEXT DEFAULT ''"},
		{"user_location", "TEXT DEFAULT ''"},
		{"user_linkedin", "TEXT DEFAULT ''"},
		{"user_github", "TEXT DEFAULT ''"},
	}
	for _, c := range cols {
		if !columnExists(db, "config", c.name) {
			slog.Info("adding missing column", "table", "config", "column", c.name)
			db.Exec("ALTER TABLE config ADD COLUMN " + c.name + " " + c.typ)
		}
	}
}

func columnExists(db *gorm.DB, table, column string) bool {
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", table, column).Scan(&count)
	return count > 0
}