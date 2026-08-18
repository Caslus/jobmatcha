package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/scanner"
	"gorm.io/gorm"
)

// ScannerService wraps the scanner engine for async job-based scans.
type ScannerService struct {
	engine *scanner.Engine
	repos  *repository.Repositories
}

// NewScannerService creates a scanner service.
func NewScannerService(db *gorm.DB, repos *repository.Repositories) *ScannerService {
	return &ScannerService{
		engine: scanner.NewEngine(db, repos),
		repos:  repos,
	}
}

// ScanResult is the public DTO for a single company scan result.
type ScanResult struct {
	CompanyName string `json:"company_name"`
	NewRoles    int    `json:"new_roles"`
	TotalRoles  int    `json:"total_roles"`
	Error       string `json:"error,omitempty"`
}

// ScanJobResponse is the public DTO for a scan job.
type ScanJobResponse struct {
	ID             uint         `json:"id"`
	Status         string       `json:"status"`
	Results        []ScanResult `json:"results,omitempty"`
	Error          string       `json:"error,omitempty"`
	DurationMS     int64        `json:"duration_ms"`
	TotalNewRoles  int          `json:"total_new_roles,omitempty"`
	TotalRoles     int          `json:"total_roles,omitempty"`
	StartedAt      *time.Time   `json:"started_at,omitempty"`
	CompletedAt    *time.Time   `json:"completed_at,omitempty"`
	CreatedAt      time.Time    `json:"created_at"`
}

// StartScan creates a scan job and runs it in the background.
// Returns the job ID immediately.
func (s *ScannerService) StartScan() (uint, error) {
	job := &model.ScanJob{
		Status: "pending",
	}
	if err := s.repos.ScanJob.Create(job); err != nil {
		return 0, err
	}

	go s.runScan(job.ID)

	return job.ID, nil
}

// runScan executes the scan in a background goroutine and updates the job.
func (s *ScannerService) runScan(jobID uint) {
	start := time.Now()

	// Mark as running
	if err := s.repos.ScanJob.UpdateStatus(jobID, "running"); err != nil {
		slog.Error("scan: failed to mark running", "job_id", jobID, "error", err)
		return
	}

	// Start the scan with a background context (outlives the HTTP request)
	results := s.engine.ScanAll(context.Background())

	// Encode results as JSON
	encoded, err := json.Marshal(resultsToDTO(results))
	if err != nil {
		s.repos.ScanJob.Fail(jobID, "encode results: "+err.Error())
		return
	}

	// Store results
	durationMS := time.Since(start).Milliseconds()
	if err := s.repos.ScanJob.Complete(jobID, string(encoded), durationMS); err != nil {
		slog.Error("scan: failed to store results", "job_id", jobID, "error", err)
		return
	}

	slog.Info("scan completed", "job_id", jobID, "companies", len(results), "took", time.Since(start).Round(time.Millisecond))
}

// GetJob returns a scan job by ID.
func (s *ScannerService) GetJob(id uint) (*ScanJobResponse, error) {
	job, err := s.repos.ScanJob.GetByID(id)
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}
	return jobToResponse(job), nil
}

// GetLatestJob returns the most recent scan job.
func (s *ScannerService) GetLatestJob() (*ScanJobResponse, error) {
	job, err := s.repos.ScanJob.GetLatest()
	if err != nil {
		return nil, err
	}
	if job == nil {
		return nil, nil
	}
	return jobToResponse(job), nil
}

func resultsToDTO(results []scanner.ScanResult) []ScanResult {
	dtos := make([]ScanResult, len(results))
	for i, r := range results {
		dtos[i] = ScanResult{CompanyName: r.CompanyName, NewRoles: r.NewRoles, TotalRoles: r.TotalRoles, Error: r.Error}
	}
	return dtos
}

func jobToResponse(job *model.ScanJob) *ScanJobResponse {
	resp := &ScanJobResponse{
		ID:          job.ID,
		Status:      job.Status,
		Error:       job.Error,
		DurationMS:  job.DurationMS,
		StartedAt:   job.StartedAt,
		CompletedAt: job.CompletedAt,
		CreatedAt:   job.CreatedAt,
	}

	if job.Results != "" {
		var results []ScanResult
		if err := json.Unmarshal([]byte(job.Results), &results); err == nil {
			resp.Results = results
			for _, r := range results {
				resp.TotalNewRoles += r.NewRoles
				resp.TotalRoles += r.TotalRoles
			}
		}
	}

	return resp
}