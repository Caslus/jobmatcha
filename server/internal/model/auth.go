package model

// ---- Request DTOs ----

type LoginRequest struct {
	Password string `json:"password"`
}

type SetupRequest struct {
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// ---- Response DTOs ----

type AuthStatusResponse struct {
	Authenticated  bool   `json:"authenticated"`
	SetupComplete  bool   `json:"setup_complete"`
	OIDCEnabled    bool   `json:"oidc_enabled"`
	OIDCProviderURL string `json:"oidc_provider_url,omitempty"`
}

type AuthTokenResponse struct {
	Token string `json:"token"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}