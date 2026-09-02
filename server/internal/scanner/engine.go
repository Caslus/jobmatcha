package scanner

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/scanner/providers"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Engine orchestrates scanning enabled boards owned by active companies.
type Engine struct {
	DB               *gorm.DB
	Repos            *repository.Repositories
	Registry         *Registry
	HTTPClient       *http.Client
	Semaphore        chan struct{}
	ProgressCallback func(completed, total int)
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

// SupportsAdapter reports whether this engine has a provider for the adapter.
func (e *Engine) SupportsAdapter(atsType string) bool {
	return e.Registry.Has(atsType)
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
	boards, err := e.Repos.CareerBoard.ListEnabledWithCompanies()
	if err != nil {
		return []ScanResult{{Error: fmt.Sprintf("load companies: %v", err)}}
	}

	results := make([]ScanResult, 0, len(boards))
	var mu sync.Mutex
	var wg sync.WaitGroup
	var completed atomic.Int32

	for _, board := range boards {
		wg.Add(1)
		e.Semaphore <- struct{}{} // acquire semaphore

		go func(b model.CareerBoard) {
			defer wg.Done()
			defer func() { <-e.Semaphore }() // release semaphore

			result := e.scanBoard(ctx, &b.Company, &b)
			mu.Lock()
			results = append(results, result)
			mu.Unlock()

			n := completed.Add(1)
			if e.ProgressCallback != nil {
				e.ProgressCallback(int(n), len(boards))
			}
		}(board)
	}

	wg.Wait()
	return results
}

// scanBoard fetches roles for a single board and upserts them under its company.
func (e *Engine) scanBoard(ctx context.Context, company *model.Company, board *model.CareerBoard) ScanResult {
	result := ScanResult{CompanyName: company.Name}
	attemptedAt := time.Now().UTC()
	if err := e.Repos.CareerBoard.RecordScanAttempt(board.ID, attemptedAt); err != nil {
		slog.Error("scanner: record attempt", "company", company.Name, "board", board.ID, "error", err)
	}

	// Resolve provider by ATS type
	atsType := board.Provider
	if atsType == "" {
		atsType = "generic"
	}

	provider, ok := e.Registry.Get(atsType)
	if !ok {
		result.Error = fmt.Sprintf("no provider for ats_type=%q", atsType)
		if err := e.Repos.CareerBoard.RecordScanFailure(board.ID, result.Error); err != nil {
			slog.Error("scanner: record provider failure", "company", company.Name, "board", board.ID, "error", err)
		}
		return result
	}

	// Fetch roles
	roles, err := provider.Fetch(ctx, company, board)
	if err != nil {
		result.Error = err.Error()
		if recordErr := e.Repos.CareerBoard.RecordScanFailure(board.ID, result.Error); recordErr != nil {
			slog.Error("scanner: record fetch failure", "company", company.Name, "board", board.ID, "error", recordErr)
		}
		return result
	}

	if err := e.Repos.CareerBoard.RecordScanSuccess(board.ID, attemptedAt); err != nil {
		slog.Error("scanner: record success", "company", company.Name, "board", board.ID, "error", err)
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

	if newCount > 0 {
		if err := e.Repos.CareerBoard.RecordNewRoleDiscovery(board.ID, attemptedAt); err != nil {
			slog.Error("scanner: record discovery", "company", company.Name, "board", board.ID, "error", err)
		}
	}

	result.NewRoles = newCount
	result.TotalRoles = len(roles)
	return result
}
