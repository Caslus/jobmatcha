package model

import "time"

// ---- Auth response DTOs ----

type AuthStatusResponse struct {
	Authenticated   bool   `json:"authenticated"`
	SetupComplete   bool   `json:"setup_complete"`
	OIDCEnabled     bool   `json:"oidc_enabled"`
	OIDCProviderURL string `json:"oidc_provider_url,omitempty"`
}

type AuthTokenResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// ---- Roles response DTOs ----

type RoleListItem struct {
	ID             uint       `json:"id"`
	CompanyID      uint       `json:"company_id"`
	CompanyName    string     `json:"company_name"`
	Title          string     `json:"title"`
	Department     string     `json:"department"`
	Location       string     `json:"location"`
	PostedAt       *time.Time `json:"posted_at"`
	RelevanceScore int        `json:"relevance_score"`
	MatchPercent   int        `json:"match_percent"`
	IsHidden       bool       `json:"is_hidden"`
	IsInterested   bool       `json:"is_interested"`
}

type PaginationInfo struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

type RoleListResponse struct {
	Data       []RoleListItem `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

type RoleDetailResponse struct {
	ID             uint       `json:"id"`
	CompanyID      uint       `json:"company_id"`
	CompanyName    string     `json:"company_name"`
	Title          string     `json:"title"`
	Department     string     `json:"department"`
	Location       string     `json:"location"`
	Description    string     `json:"description"`
	URL            string     `json:"url"`
	PostedAt       *time.Time `json:"posted_at"`
	RelevanceScore int        `json:"relevance_score"`
	MatchPercent   int        `json:"match_percent"`
	IsHidden       bool       `json:"is_hidden"`
	IsInterested   bool       `json:"is_interested"`
}