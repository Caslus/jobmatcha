package model

import "time"

// ScanJob tracks a background scan run.
// Status: pending → running → completed | failed
type ScanJob struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	Status             string     `gorm:"size:20;not null;default:pending" json:"status"`
	Results            string     `gorm:"type:text" json:"results,omitempty"` // JSON-encoded []ScanResult
	Error              string     `gorm:"type:text" json:"error,omitempty"`
	DurationMS         int64      `gorm:"not null;default:0" json:"duration_ms"`
	TotalCompanies     int        `gorm:"not null;default:0" json:"total_companies"`
	CompletedCompanies int        `gorm:"not null;default:0" json:"completed_companies"`
	StartedAt          *time.Time `json:"started_at,omitempty"`
	CompletedAt        *time.Time `json:"completed_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

func (ScanJob) TableName() string { return "scan_jobs" }

// ScanResult is the public DTO for a single company scan result.
type ScanResult struct {
	CompanyName string `json:"company_name"`
	NewRoles    int    `json:"new_roles"`
	TotalRoles  int    `json:"total_roles"`
	Error       string `json:"error,omitempty"`
}

// ScanJobResponse is the public DTO returned by scan endpoints.
type ScanJobResponse struct {
	ID                 uint         `json:"id"`
	Status             string       `json:"status"`
	Results            []ScanResult `json:"results,omitempty"`
	Error              string       `json:"error,omitempty"`
	DurationMS         int64        `json:"duration_ms"`
	TotalCompanies     int          `json:"total_companies"`
	CompletedCompanies int          `json:"completed_companies"`
	TotalNewRoles      int          `json:"total_new_roles"`
	TotalRoles         int          `json:"total_roles"`
	StartedAt          *time.Time   `json:"started_at,omitempty"`
	CompletedAt        *time.Time   `json:"completed_at,omitempty"`
	CreatedAt          time.Time    `json:"created_at"`
}
