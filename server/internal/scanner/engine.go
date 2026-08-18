package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/scanner/providers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Engine orchestrates scanning all active companies through their ATS providers.
type Engine struct {
	DB         *gorm.DB
	Repos      *repository.Repositories
	Registry   *Registry
	HTTPClient *http.Client
	Semaphore  chan struct{}
}

// NewEngine creates a scanner engine with the built-in providers registered.
func NewEngine(db *gorm.DB, repos *repository.Repositories) *Engine {
	e := &Engine{
		DB:         db,
		Repos:      repos,
		Registry:   NewRegistry(),
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
		Semaphore:  make(chan struct{}, 5), // max 5 concurrent requests
	}
	e.registerBuiltinProviders()
	return e
}

func (e *Engine) registerBuiltinProviders() {
	providers := []Provider{
		&providers.Workable{HTTPClient: e.HTTPClient},
		&providers.Greenhouse{HTTPClient: e.HTTPClient},
	}
	for _, p := range providers {
		e.Registry.Register(p)
	}
}

// ScanResult summarizes a single company scan.
type ScanResult struct {
	CompanyName string `json:"company_name"`
	NewRoles    int    `json:"new_roles"`
	TotalRoles  int    `json:"total_roles"`
	Error       string `json:"error,omitempty"`
}

// ScanAll scans all active companies concurrently and returns per-company results.
func (e *Engine) ScanAll(ctx context.Context) []ScanResult {
	companies, err := e.Repos.Company.ListActive()
	if err != nil {
		return []ScanResult{{Error: fmt.Sprintf("load companies: %v", err)}}
	}

	results := make([]ScanResult, 0, len(companies))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, company := range companies {
		wg.Add(1)
		e.Semaphore <- struct{}{} // acquire semaphore

		go func(c model.Company) {
			defer wg.Done()
			defer func() { <-e.Semaphore }() // release semaphore

			result := e.scanCompany(ctx, &c)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()
		}(company)
	}

	wg.Wait()
	return results
}

// scanCompany fetches roles for a single company via its provider and upserts them.
func (e *Engine) scanCompany(ctx context.Context, company *model.Company) ScanResult {
	result := ScanResult{CompanyName: company.Name}

	// Resolve provider by ATS type
	atsType := company.ATSType
	if atsType == "" {
		atsType = "generic"
	}

	provider, ok := e.Registry.Get(atsType)
	if !ok {
		result.Error = fmt.Sprintf("no provider for ats_type=%q", atsType)
		return result
	}

	// Fetch roles
	roles, err := provider.Fetch(ctx, company)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	if len(roles) == 0 {
		result.Error = "no jobs found"
		return result
	}

	// Upsert roles — count new vs existing
	newCount := 0
	for _, role := range roles {
		role.CompanyID = company.ID
		role.URLHash = URLHash(company.ID, role.URL)
		if role.Status == "" {
			role.Status = "seen"
		}

		// Try insert — OnConflict DoNothing means RowsAffected=1 only for true inserts
		dup := e.DB.Clauses(clause.OnConflict{DoNothing: true}).Create(role)
		if dup.Error != nil {
			slog.Error("scanner: insert role", "company", company.Name, "title", role.Title, "error", dup.Error)
			continue
		}
		if dup.RowsAffected > 0 {
			newCount++
		} else {
			// Role already exists — refresh data fields silently
			e.DB.Model(&model.Role{}).Where("url_hash = ?", role.URLHash).Updates(map[string]interface{}{
				"title":       role.Title,
				"description": role.Description,
				"location":    role.Location,
				"department":  role.Department,
				"posted_at":   role.PostedAt,
			})
		}
	}

	// Update last_scanned_at
	e.DB.Model(&model.Company{}).Where("id = ?", company.ID).
		Update("last_scanned_at", time.Now())

	result.NewRoles = newCount
	result.TotalRoles = len(roles)
	return result
}