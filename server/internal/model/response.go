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

// ---- Match details ----

type MatchDetails struct {
	IncludeScore  int     `json:"include_score"`
	BonusScore    int     `json:"bonus_score"`
	TotalScore    int     `json:"total_score"`
	RecencyFactor float64 `json:"recency_factor"`
	AdjustedScore float64 `json:"adjusted_score"`
	MatchedKw     int     `json:"matched_keywords"`
	TotalKw       int     `json:"total_keywords"`
	Percent       int     `json:"percent"`
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
	ID             uint         `json:"id"`
	CompanyID      uint         `json:"company_id"`
	CompanyName    string       `json:"company_name"`
	Title          string       `json:"title"`
	Department     string       `json:"department"`
	Location       string       `json:"location"`
	Description    string       `json:"description"`
	URL            string       `json:"url"`
	PostedAt       *time.Time   `json:"posted_at"`
	RelevanceScore int          `json:"relevance_score"`
	MatchPercent   int          `json:"match_percent"`
	MatchReasons   []string     `json:"match_reasons,omitempty"`
	MatchDetails   *MatchDetails `json:"match_details,omitempty"`
	IsHidden       bool         `json:"is_hidden"`
	IsInterested   bool         `json:"is_interested"`
}