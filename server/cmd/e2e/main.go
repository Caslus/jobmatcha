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

func (noOpScanner) StartScan() (uint, error)                    { return 0, nil }
func (noOpScanner) GetJob(uint) (*model.ScanJobResponse, error) { return fixtureScanResponse(), nil }
func (noOpScanner) GetLatestJob() (*model.ScanJobResponse, error) {
	return fixtureScanResponse(), nil
}
func (noOpScanner) SupportsAdapter(string) bool { return false }

func fixtureScanResponse() *model.ScanJobResponse {
	completedAt := time.Now().Add(-30 * time.Minute)
	return &model.ScanJobResponse{
		ID: 1, Status: "completed", DurationMS: 4280,
		TotalCompanies: 2, CompletedCompanies: 2, CompletedAt: &completedAt,
		TotalNewRoles: 12, TotalRoles: 18,
		Results: []model.ScanResult{
			{CompanyName: "SmartNews", NewRoles: 7, TotalRoles: 9},
			{CompanyName: "Mercari", NewRoles: 5, TotalRoles: 9},
		},
	}
}

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
			companies := []model.Company{
				{Name: "SmartNews", CareersURL: "https://careers.smartnews.com/", Active: false},
				{Name: "Mercari", CareersURL: "https://careers.mercari.com/", Active: false},
			}
			for i := range companies {
				if err := db.Where(model.Company{Name: companies[i].Name}).FirstOrCreate(&companies[i]).Error; err != nil {
					return err
				}
			}
			return nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}
	defer application.Close()
	if err := application.Repos.Config.UpdateMap(context.Background(), map[string]interface{}{
		"setup_complete": true,
		// A screenshot-friendly representative subset of the production resume
		// parser output, with Tokyo/Japan applied as the explicit target.
		"include_keywords": model.StringSlice{
			"software engineer", "backend", "frontend", "full stack", "java", "spring", "go", "react",
			"typescript", "javascript", "node.js", "react native", "rest api", "angular", "html", "css",
			"oracle", "sql", "postgresql", "mysql", "mongodb", "aws", "docker", "github actions", "observability", "japanese",
		},
		"exclude_keywords": model.StringSlice{
			"senior", "lead", "principal", "staff", "manager", "director", "head", "specialist", "recruiter", "intern", "シニア", "リード",
		},
		"location_keywords": model.StringSlice{"remote", "hybrid", "tokyo", "osaka", "sapporo", "fukuoka", "kyoto"},
		"work_types":        model.StringSlice{},
		"max_days_old":      14,
		"ai_provider":       "openrouter",
		"ai_enabled":        true,
	}); err != nil {
		log.Fatal(err)
	}
	postedAt := time.Now().Add(-90 * time.Minute)
	// Public listing metadata captured from a Tokyo scan on 2026-08-28. The
	// descriptions are original summaries, keeping the fixture reproducible
	// without copying live job-posting text.
	roles := []model.Role{
		{CompanyID: 1, URLHash: "smartnews-ads-backend", URL: "https://careers.smartnews.com/", Title: "Software Engineer, Ads Backend", Department: "Engineering", Location: "Shibuya, Tokyo, Japan", Description: "## Build advertising systems at SmartNews\n\nSmartNews is hiring a Software Engineer for its Ads Backend team in Tokyo. This role is a strong fit for an engineer who enjoys building dependable product infrastructure across backend services and web-facing systems.\n\nYou will work with product and platform partners to design APIs, deliver services, and improve the systems that support advertising experiences. The team values clear technical communication, thoughtful trade-offs, and ownership from design through production.\n\n### What you will bring\n\n- Production software engineering experience across backend and full stack work.\n- Hands-on React and TypeScript experience, plus Go, Java, Spring, or Node.js.\n- Familiarity with PostgreSQL, AWS, Docker, and CI/CD workflows.\n- A collaborative approach to shipping reliable systems for customers and internal teams.\n\n### What you will do\n\n- Build and operate backend APIs and services for advertising products.\n- Partner with frontend engineers on full stack features and high-quality integrations.\n- Improve delivery practices, observability, and the developer experience.\n- Help the team make pragmatic architecture decisions as the product evolves.", DescriptionFormat: "markdown", PostedAt: &postedAt},
		{CompanyID: 1, URLHash: "smartnews-webtech-frontend", URL: "https://careers.smartnews.com/", Title: "フロントエンドエンジニア / Software Engineer, Webtech Frontend", Department: "Engineering", Location: "Shibuya, Tokyo, Japan", Description: "Build modern web experiences for SmartNews product teams in Tokyo.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-3 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-experimentation-system", URL: "https://careers.mercari.com/", Title: "Software Engineer - Experimentation System - Mercari", Department: "Engineering", Location: "Minato City, Tokyo, Japan", Description: "Develop experimentation systems with Java, Go, TypeScript, Node.js, PostgreSQL, AWS, Docker, and GitHub Actions. Partner with product teams to make decisions with confidence and improve reliable feature delivery.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-6 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-backend-shops", URL: "https://careers.mercari.com/", Title: "Product Engineer, Backend(Shops) - Mercari", Department: "Engineering", Location: "Minato City, Tokyo, Japan", Description: "Build backend services with Java, Spring, Go, PostgreSQL, AWS, Docker, and GitHub Actions for Mercari Shops products in Tokyo. Work closely with frontend teams on dependable full stack customer experiences.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-5 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-ml-operations", URL: "https://careers.mercari.com/", Title: "ML Operations Engineer (AI/LLM) - Mercari", Department: "Engineering", Location: "Minato City, Tokyo, Japan", Description: "Support AI and LLM operations with Java, Spring, Go, PostgreSQL, AWS, Docker, and GitHub Actions for Mercari engineering teams. Improve observability and production delivery workflows.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-90 * time.Minute))},
		{CompanyID: 2, URLHash: "mercari-talent-development", URL: "https://careers.mercari.com/", Title: "Talent Development Specialist - Mercari", Department: "People", Location: "Minato City, Tokyo, Japan", Description: "Support talent development programs at Mercari.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-2 * 24 * time.Hour))},
		{CompanyID: 1, URLHash: "smartnews-business-development", URL: "https://careers.smartnews.com/", Title: "Business Development, New Era", Department: "Business", Location: "Shibuya, Tokyo, Japan", Description: "Develop new business opportunities for SmartNews in Tokyo.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-7 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-culture-relations", URL: "https://careers.mercari.com/", Title: "Culture Relations Lead - Mercari", Department: "People", Location: "Minato City, Tokyo, Japan", Description: "Lead culture and relations initiatives at Mercari.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-8 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-corporate-planning", URL: "https://careers.mercari.com/", Title: "Corporate Planning Specialist / 経営企画 - Mercari Marketplace", Department: "Corporate", Location: "Minato City, Tokyo, Japan", Description: "Support corporate planning for Mercari Marketplace.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-8 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-treasury", URL: "https://careers.mercari.com/", Title: "Treasury Specialist / トレジャリースペシャリスト - Mercari", Department: "Finance", Location: "Minato City, Tokyo, Japan", Description: "Support treasury operations at Mercari.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-9 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "mercari-marketing", URL: "https://careers.mercari.com/", Title: "Marketing Specialist / マーケティングスペシャリスト - Mercari / Merpay", Department: "Marketing", Location: "Minato City, Tokyo, Japan", Description: "Support marketing work across Mercari and Merpay.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-9 * 24 * time.Hour))},
		{CompanyID: 1, URLHash: "smartnews-account-manager", URL: "https://careers.smartnews.com/", Title: "アカウントマネージャー(IS4)", Department: "Sales", Location: "Shibuya, Tokyo, Japan", Description: "Manage advertising customer relationships at SmartNews.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-10 * 24 * time.Hour))},
		// The entries below are fixture-only score-calibration variants. They are
		// intentionally not represented as live listings in project documentation.
		{CompanyID: 1, URLHash: "fixture-backend-platform", URL: "https://careers.smartnews.com/", Title: "Backend Engineer, Platform", Department: "Engineering", Location: "Shibuya, Tokyo, Japan", Description: "Build Go and Java services with Spring, TypeScript, Node.js, PostgreSQL, AWS, Docker, GitHub Actions, Linux, Bash, and Grafana. Improve observability for platform teams.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-2 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "fixture-full-stack", URL: "https://careers.mercari.com/", Title: "Full Stack Engineer, Developer Experience", Department: "Engineering", Location: "Minato City, Tokyo, Japan", Description: "Ship React and TypeScript experiences with Java, Spring, Node.js, REST APIs, PostgreSQL, AWS, Docker, and GitHub Actions. Improve the developer workflow for full stack product teams.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-4 * 24 * time.Hour))},
		{CompanyID: 1, URLHash: "fixture-frontend-developer", URL: "https://careers.smartnews.com/", Title: "Frontend Developer, Web Platform", Department: "Engineering", Location: "Shibuya, Tokyo, Japan", Description: "Build React, TypeScript, CSS, Node.js, PostgreSQL, AWS, and Docker interfaces for Tokyo product teams. Collaborate on reliable full stack web experiences.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-3 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "fixture-application-engineer", URL: "https://careers.mercari.com/", Title: "Application Engineer, Payments", Department: "Engineering", Location: "Minato City, Tokyo, Japan", Description: "Improve Java, Spring, Go, SQL, REST API, PostgreSQL, AWS, Docker, and GitHub Actions integrations for customer applications. Help teams ship reliable payment experiences.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-7 * 24 * time.Hour))},
		{CompanyID: 1, URLHash: "fixture-reliability-engineer", URL: "https://careers.smartnews.com/", Title: "Reliability Engineer", Department: "Engineering", Location: "Shibuya, Tokyo, Japan", Description: "Improve observability with Dynatrace, Grafana, and incident management practices.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-8 * 24 * time.Hour))},
		{CompanyID: 2, URLHash: "fixture-mobile-engineer", URL: "https://careers.mercari.com/", Title: "Mobile Engineer", Department: "Engineering", Location: "Minato City, Tokyo, Japan", Description: "Build React Native product experiences for customers in Japan.", DescriptionFormat: "plain", PostedAt: timePtr(postedAt.Add(-11 * 24 * time.Hour))},
	}
	if err := application.DB.Create(&roles).Error; err != nil {
		log.Fatal(err)
	}
	log.Printf("e2e fixture listening on :%s", *port)
	log.Fatal(http.ListenAndServe(":"+*port, application.Router))
}

func timePtr(value time.Time) *time.Time { return &value }
