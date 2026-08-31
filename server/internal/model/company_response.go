package model

import "time"

// CompanyListItem is the management view of one registered job source.
type CompanyListItem struct {
	ID                     uint       `json:"id"`
	Name                   string     `json:"name"`
	Location               string     `json:"location"`
	ATSType                string     `json:"ats_type"`
	Active                 bool       `json:"active"`
	RoleCount              int64      `json:"role_count"`
	AdapterStatus          string     `json:"adapter_status"`
	FreshnessStatus        string     `json:"freshness_status"`
	LastScanAttemptAt      *time.Time `json:"last_scan_attempt_at"`
	LastSuccessfulScanAt   *time.Time `json:"last_successful_scan_at"`
	LastScanFailureDetail  *string    `json:"last_scan_failure_detail"`
	LastNewRoleDiscoveryAt *time.Time `json:"last_new_role_discovery_at"`
}

type CompanyListResponse struct {
	Data []CompanyListItem `json:"data"`
}

type CompanyActiveUpdateRequest struct {
	Active *bool `json:"active" binding:"required"`
}

type CompanyBulkActiveUpdateRequest struct {
	CompanyIDs []uint `json:"company_ids" binding:"required,min=1"`
	Active     *bool  `json:"active" binding:"required"`
}
