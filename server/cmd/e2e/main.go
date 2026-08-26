// Command e2e runs a local, hermetic Jobmatcha fixture for browser tests.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"path/filepath"

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
	log.Printf("e2e fixture listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, application.Router))
}
