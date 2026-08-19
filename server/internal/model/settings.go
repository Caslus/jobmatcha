package model

// SettingsResponse returns the user's preferences (password hash excluded).
type SettingsResponse struct {
	IncludeKeywords  StringSlice `json:"include_keywords"`
	ExcludeKeywords  StringSlice `json:"exclude_keywords"`
	LocationKeywords StringSlice `json:"location_keywords"`
	WorkTypes        StringSlice `json:"work_types"`
	EmploymentType   string      `json:"employment_type"`
	MaxDaysOld       int         `json:"max_days_old"`
}

// SettingsUpdateRequest is the request body for PUT /api/settings.
type SettingsUpdateRequest struct {
	IncludeKeywords  *StringSlice `json:"include_keywords,omitempty"`
	ExcludeKeywords  *StringSlice `json:"exclude_keywords,omitempty"`
	LocationKeywords *StringSlice `json:"location_keywords,omitempty"`
	WorkTypes        *StringSlice `json:"work_types,omitempty"`
	EmploymentType   *string      `json:"employment_type,omitempty"`
	MaxDaysOld       *int         `json:"max_days_old,omitempty"`
}