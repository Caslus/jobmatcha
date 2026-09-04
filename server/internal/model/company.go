package model

import "time"

type Company struct {
	ID                     uint          `gorm:"primaryKey" json:"id"`
	Name                   string        `gorm:"uniqueIndex;not null;size:255" json:"name"`
	CareersURL             string        `gorm:"not null;size:1024" json:"careers_url"`
	ATSType                string        `gorm:"size:50" json:"ats_type"`
	ATSSlug                string        `gorm:"size:255" json:"ats_slug"`
	Active                 bool          `gorm:"not null;default:true" json:"active"`
	LastScannedAt          *time.Time    `json:"last_scanned_at"`
	LastScanAttemptAt      *time.Time    `json:"last_scan_attempt_at"`
	LastSuccessfulScanAt   *time.Time    `json:"last_successful_scan_at"`
	LastScanFailureDetail  *string       `gorm:"type:text" json:"last_scan_failure_detail"`
	LastNewRoleDiscoveryAt *time.Time    `json:"last_new_role_discovery_at"`
	CareerBoards           []CareerBoard `gorm:"foreignKey:CompanyID" json:"career_boards,omitempty"`
	CreatedAt              time.Time     `json:"created_at"`
	UpdatedAt              time.Time     `json:"updated_at"`
}

func (Company) TableName() string { return "companies" }
