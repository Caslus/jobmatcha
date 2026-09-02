package model

import "time"

// CompanyListItem is the company-level management summary.
type CompanyListItem struct {
	ID                     uint                  `json:"id"`
	Name                   string                `json:"name"`
	Location               string                `json:"location"`
	Active                 bool                  `json:"active"`
	BoardCount             int                   `json:"board_count"`
	RoleCount              int64                 `json:"role_count"`
	FreshnessStatus        string                `json:"freshness_status"`
	LastScanAttemptAt      *time.Time            `json:"last_scan_attempt_at"`
	LastNewRoleDiscoveryAt *time.Time            `json:"last_new_role_discovery_at"`
	CareerBoards           []CareerBoardListItem `json:"career_boards"`
}

// CareerBoardListItem is one independently managed source for a company.
type CareerBoardListItem struct {
	ID                     uint       `json:"id"`
	Provider               string     `json:"provider"`
	BoardIdentifier        string     `json:"board_identifier"`
	CanonicalURL           string     `json:"canonical_url"`
	Active                 bool       `json:"active"`
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

type CareerBoardActiveUpdateRequest struct {
	Active *bool `json:"active" binding:"required"`
}

type CareerBoardDiscoveryRequest struct {
	CareersURL string `json:"careers_url" binding:"required"`
}

type CareerBoardDiscoveryCandidate struct {
	Provider         string   `json:"provider"`
	BoardIdentifier  string   `json:"board_identifier"`
	CanonicalURL     string   `json:"canonical_url"`
	EvidenceURLs     []string `json:"evidence_urls"`
	ValidationStatus string   `json:"validation_status"`
	ValidationError  string   `json:"validation_error,omitempty"`
}

type CareerBoardDiscoveryResponse struct {
	Candidates             []CareerBoardDiscoveryCandidate `json:"candidates"`
	EmployerNameSuggestion string                          `json:"employer_name_suggestion"`
	Incomplete             bool                            `json:"incomplete"`
}

type CareerBoardRegistration struct {
	CompanyName     string `json:"company_name" binding:"required"`
	CareersURL      string `json:"careers_url" binding:"required"`
	Location        string `json:"location"`
	Region          string `json:"region"`
	Provider        string `json:"provider" binding:"required"`
	BoardIdentifier string `json:"board_identifier" binding:"required"`
	CanonicalURL    string `json:"canonical_url" binding:"required"`
	Separate        bool   `json:"separate"`
}

type CareerBoardRegistrationRequest struct {
	Candidates []CareerBoardRegistration `json:"candidates"`
}
