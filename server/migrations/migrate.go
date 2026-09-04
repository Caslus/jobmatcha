package migrations

import (
	"fmt"
	"log/slog"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	if err := backfillResumeColumns(db); err != nil {
		return err
	}

	if err := db.AutoMigrate(
		&model.Company{},
		&model.CareerBoard{},
		&model.Role{},
		&model.Config{},
		&model.Session{},
		&model.ScanJob{},
		&model.Resume{},
		&model.TailoredResume{},
	); err != nil {
		return err
	}
	if err := dropRetiredCompanyColumns(db); err != nil {
		return err
	}

	if err := backfillCareerBoards(db); err != nil {
		return err
	}

	// Backfill columns that may not exist on older databases
	// (GORM AutoMigrate doesn't always add columns to existing SQLite tables)
	return backfillConfigColumns(db)
}

// dropRetiredCompanyColumns removes company metadata that duplicated the
// authoritative location stored on each scanned role.
func dropRetiredCompanyColumns(db *gorm.DB) error {
	for _, column := range []string{"region", "location"} {
		if db.Migrator().HasColumn(&model.Company{}, column) {
			if err := db.Migrator().DropColumn(&model.Company{}, column); err != nil {
				return fmt.Errorf("dropping companies.%s: %w", column, err)
			}
		}
	}
	return nil
}

// backfillCareerBoards preserves legacy single-board sources after upgrading.
// It is idempotent because provider and board_identifier have a unique key.
func backfillCareerBoards(db *gorm.DB) error {
	var companies []model.Company
	if err := db.Where("ats_type <> '' AND ats_slug <> ''").Find(&companies).Error; err != nil {
		return fmt.Errorf("loading legacy company boards: %w", err)
	}
	for _, company := range companies {
		board := model.CareerBoard{
			CompanyID: company.ID, Provider: company.ATSType, BoardIdentifier: company.ATSSlug,
			CanonicalURL: company.CareersURL, Active: company.Active,
			LastScannedAt: company.LastScannedAt, LastScanAttemptAt: company.LastScanAttemptAt,
			LastSuccessfulScanAt: company.LastSuccessfulScanAt, LastScanFailureDetail: company.LastScanFailureDetail,
			LastNewRoleDiscoveryAt: company.LastNewRoleDiscoveryAt,
		}
		if err := db.Where("provider = ? AND board_identifier = ?", board.Provider, board.BoardIdentifier).FirstOrCreate(&board).Error; err != nil {
			return fmt.Errorf("backfilling company %d career board: %w", company.ID, err)
		}
	}
	return nil
}

// SQLite cannot add a NOT NULL column without a default value to a table that
// already contains rows. Add the structured document column ourselves before
// AutoMigrate so older databases receive an empty JSON document safely.
func backfillResumeColumns(db *gorm.DB) error {
	if !db.Migrator().HasTable(&model.Resume{}) || columnExists(db, "resumes", "document") {
		return nil
	}

	slog.Info("adding missing column", "table", "resumes", "column", "document")
	if err := db.Exec("ALTER TABLE resumes ADD COLUMN document TEXT NOT NULL DEFAULT '{}'").Error; err != nil {
		return fmt.Errorf("adding resumes.document: %w", err)
	}
	return nil
}

func backfillConfigColumns(db *gorm.DB) error {
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
			if err := db.Exec("ALTER TABLE config ADD COLUMN " + c.name + " " + c.typ).Error; err != nil {
				return fmt.Errorf("adding config.%s: %w", c.name, err)
			}
		}
	}
	return nil
}

func columnExists(db *gorm.DB, table, column string) bool {
	var count int64
	db.Raw("SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?", table, column).Scan(&count)
	return count > 0
}
