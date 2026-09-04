package migrations

import (
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type legacyCompany struct {
	ID         uint   `gorm:"primaryKey"`
	Name       string `gorm:"uniqueIndex;not null"`
	CareersURL string `gorm:"not null"`
	Region     string `gorm:"not null;default:JP"`
	Location   string
}

func (legacyCompany) TableName() string { return "companies" }

func TestSeedCompaniesUsesOnlySupportedAdapters(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	if err := SeedCompanies(db); err != nil {
		t.Fatalf("seed companies: %v", err)
	}

	var companies []model.Company
	if err := db.Preload("CareerBoards").Order("name").Find(&companies).Error; err != nil {
		t.Fatalf("load seeded companies: %v", err)
	}
	if len(companies) == 0 {
		t.Fatal("expected the seed to create companies")
	}

	for _, company := range companies {
		if company.ATSType != "workable" && company.ATSType != "greenhouse" {
			t.Errorf("%s uses unsupported adapter %q", company.Name, company.ATSType)
		}
		if company.ATSSlug == "" {
			t.Errorf("%s has no board identifier", company.Name)
		}
		if len(company.CareerBoards) != 1 {
			t.Errorf("%s boards = %d, want 1", company.Name, len(company.CareerBoards))
		}
	}

	if err := SeedCompanies(db); err != nil {
		t.Fatalf("seed companies a second time: %v", err)
	}

	var repeatedCount int64
	if err := db.Model(&model.Company{}).Count(&repeatedCount).Error; err != nil {
		t.Fatalf("count companies after repeated seed: %v", err)
	}
	if repeatedCount != int64(len(companies)) {
		t.Errorf("repeated seed created duplicate companies: got %d, want %d", repeatedCount, len(companies))
	}
}

func TestMigrateDropsRetiredCompanyMetadata(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&legacyCompany{}); err != nil {
		t.Fatalf("create legacy companies table: %v", err)
	}
	if err := db.Create(&legacyCompany{Name: "Legacy", CareersURL: "https://legacy.example", Region: "JP", Location: "Tokyo"}).Error; err != nil {
		t.Fatalf("create legacy company: %v", err)
	}

	if err := Migrate(db); err != nil {
		t.Fatalf("migrate legacy database: %v", err)
	}
	for _, column := range []string{"region", "location"} {
		if db.Migrator().HasColumn(&model.Company{}, column) {
			t.Errorf("companies.%s still exists after migration", column)
		}
	}

	var company model.Company
	if err := db.Where("name = ?", "Legacy").First(&company).Error; err != nil {
		t.Fatalf("load migrated company: %v", err)
	}
}
