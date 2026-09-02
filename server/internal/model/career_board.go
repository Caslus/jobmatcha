package model

import "time"

// BoardIdentity is a provider-normalized external board reference.
type BoardIdentity struct {
	Provider        string `json:"provider"`
	BoardIdentifier string `json:"board_identifier"`
	CanonicalURL    string `json:"canonical_url"`
}

// CareerBoard is an independently managed external job-board source for a
// company. Roles remain owned by the company so that changing a source does
// not change existing role ownership.
type CareerBoard struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	CompanyID              uint       `gorm:"not null;index" json:"company_id"`
	Company                Company    `gorm:"foreignKey:CompanyID" json:"-"`
	Provider               string     `gorm:"not null;size:50;uniqueIndex:idx_career_boards_provider_identifier" json:"provider"`
	BoardIdentifier        string     `gorm:"not null;size:255;uniqueIndex:idx_career_boards_provider_identifier" json:"board_identifier"`
	CanonicalURL           string     `gorm:"not null;size:1024" json:"canonical_url"`
	Active                 bool       `gorm:"not null;default:true" json:"active"`
	LastScannedAt          *time.Time `json:"last_scanned_at"`
	LastScanAttemptAt      *time.Time `json:"last_scan_attempt_at"`
	LastSuccessfulScanAt   *time.Time `json:"last_successful_scan_at"`
	LastScanFailureDetail  *string    `gorm:"type:text" json:"last_scan_failure_detail"`
	LastNewRoleDiscoveryAt *time.Time `json:"last_new_role_discovery_at"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
}

func (CareerBoard) TableName() string { return "career_boards" }
