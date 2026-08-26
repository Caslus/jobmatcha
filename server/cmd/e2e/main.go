// Command e2e runs a local, hermetic Jobmatcha fixture for browser tests.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/caslus/jobmatcha/internal/app"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"gorm.io/gorm"
)

type noOpScanner struct{}

func (noOpScanner) StartScan() (uint, error)                      { return 0, nil }
func (noOpScanner) GetJob(uint) (*model.ScanJobResponse, error)   { return nil, nil }
func (noOpScanner) GetLatestJob() (*model.ScanJobResponse, error) { return nil, nil }

func main() {
	port := flag.String("port", "8182", "HTTP port")
	staticDir := flag.String("static-dir", "web/dist/client", "built SPA directory")
	workDir := flag.String("work-dir", "", "fixture directory (required)")
	flag.Parse()
	if *workDir == "" {
		log.Fatal("-work-dir is required")
	}
	for _, name := range []string{"fixture.db", "fixture.db-shm", "fixture.db-wal", "bootstrap-password"} {
		if err := os.Remove(filepath.Join(*workDir, name)); err != nil && !os.IsNotExist(err) {
			log.Fatal(err)
		}
	}
	if err := os.MkdirAll(*workDir, 0o700); err != nil {
		log.Fatal(err)
	}

	application, err := app.New(context.Background(), app.Options{
		DBPath:                filepath.Join(*workDir, "fixture.db"),
		BootstrapPasswordFile: filepath.Join(*workDir, "bootstrap-password"),
		StaticDir:             *staticDir, DisableScheduler: true,
		NewScanner: func(*gorm.DB, *repository.Repositories) service.Scanner { return noOpScanner{} },
		SeedCompanies: func(db *gorm.DB) error {
			company := model.Company{Name: "Fixture Co", CareersURL: "https://fixture.invalid", Active: false}
			return db.Where(model.Company{Name: company.Name}).FirstOrCreate(&company).Error
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()
	if err := application.Repos.Config.UpdateMap(context.Background(), map[string]interface{}{
		"setup_complete":    true,
		"include_keywords":  model.StringSlice{"go", "typescript"},
		"exclude_keywords":  model.StringSlice{},
		"location_keywords": model.StringSlice{"remote"},
		"work_types":        model.StringSlice{"remote"},
	}); err != nil {
		log.Fatal(err)
	}
	postedAt := time.Now().Add(-time.Hour)
	role := model.Role{
		CompanyID:         1,
		URLHash:           "fixture-role",
		URL:               "https://fixture.invalid/jobs/software-engineer",
		Title:             "Software Engineer",
		Location:          "Remote",
		Description:       "Build reliable software.",
		DescriptionFormat: "plain",
		PostedAt:          &postedAt,
	}
	if err := application.DB.Create(&role).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("e2e fixture listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, application.Router))
}
