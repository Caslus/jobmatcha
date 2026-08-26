package model

// ---- AI validation ----

type AIValidateKeyRequest struct {
	Provider string `json:"provider"`
	APIKey   string `json:"api_key"`
}

type AIValidateKeyResponse struct {
	Valid  bool   `json:"valid"`
	Error  string `json:"error,omitempty"`
	Models int    `json:"models,omitempty"`
}

// ---- Resume parsing ----

// ParseResumeResponse is the structured data extracted from a resume by the LLM.
type ParseResumeResponse struct {
	Name                      string             `json:"name"`
	Email                     string             `json:"email"`
	Location                  string             `json:"location"`
	LinkedinURL               string             `json:"linkedin_url,omitempty"`
	GithubURL                 string             `json:"github_url,omitempty"`
	SuggestedInclude          []string           `json:"suggested_include"`
	SuggestedExclude          []string           `json:"suggested_exclude"`
	SuggestedWorkTypes        []string           `json:"suggested_work_types"`
	SuggestedLocationKeywords []string           `json:"suggested_location_keywords"`
	Resume                    ResumeInfoResponse `json:"resume"`
}

// ---- AI settings (exposed via GET/PUT /api/settings/ai) ----

type AIInfoResponse struct {
	Provider     string `json:"provider"`
	Enabled      bool   `json:"enabled"`
	HasAPIKey    bool   `json:"has_api_key"` // true if a key is stored (never returns the key itself)
	UserName     string `json:"user_name"`
	UserEmail    string `json:"user_email"`
	UserLocation string `json:"user_location"`
	UserLinkedin string `json:"user_linkedin"`
	UserGithub   string `json:"user_github"`
}

type AIUpdateRequest struct {
	Provider     *string `json:"provider,omitempty"`
	APIKey       *string `json:"api_key,omitempty"` // set to update, null/omit to keep existing
	Enabled      *bool   `json:"enabled,omitempty"`
	UserName     *string `json:"user_name,omitempty"`
	UserEmail    *string `json:"user_email,omitempty"`
	UserLocation *string `json:"user_location,omitempty"`
	UserLinkedin *string `json:"user_linkedin,omitempty"`
	UserGithub   *string `json:"user_github,omitempty"`
}

// ---- Onboarding completion ----

// OnboardingCompleteRequest is the final step of the wizard.
type OnboardingCompleteRequest struct {
	// User profile
	UserName     string `json:"user_name"`
	UserEmail    string `json:"user_email"`
	UserLocation string `json:"user_location"`
	UserLinkedin string `json:"user_linkedin,omitempty"`
	UserGithub   string `json:"user_github,omitempty"`

	// Relevance preferences
	IncludeKeywords  StringSlice `json:"include_keywords"`
	ExcludeKeywords  StringSlice `json:"exclude_keywords"`
	LocationKeywords StringSlice `json:"location_keywords"`
	WorkTypes        StringSlice `json:"work_types"`
	MaxDaysOld       int         `json:"max_days_old"`

	// Scan settings
	ScanEnabled  bool   `json:"scan_enabled"`
	ScanCronExpr string `json:"scan_cron_expr"`
	ScanTimezone string `json:"scan_timezone"`
}
