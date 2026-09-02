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
