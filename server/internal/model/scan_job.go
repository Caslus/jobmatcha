package model

import "time"

// ScanJob tracks a background scan run.
// Status: pending → running → completed | failed
type ScanJob struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Status      string     `gorm:"size:20;not null;default:pending" json:"status"`
	Results     string     `gorm:"type:text" json:"results,omitempty"` // JSON-encoded []ScanResult
	Error       string     `gorm:"type:text" json:"error,omitempty"`
	DurationMS  int64      `gorm:"not null;default:0" json:"duration_ms"`
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (ScanJob) TableName() string { return "scan_jobs" }
