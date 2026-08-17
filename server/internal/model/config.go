package model

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type StringSlice []string

func (s StringSlice) Value() (driver.Value, error) {
	return json.Marshal(s)
}

func (s *StringSlice) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to scan StringSlice: expected []byte, got %T", value)
	}
	return json.Unmarshal(bytes, s)
}

type Config struct {
	ID               uint   `gorm:"primaryKey;check:id=1" json:"-"`
	PasswordHash     string `gorm:"not null;default:''" json:"-"`
	SetupComplete    bool   `gorm:"not null;default:false" json:"setup_complete"`

	// OIDC (optional)
	OIDCEnabled      bool   `gorm:"not null;default:false" json:"oidc_enabled"`
	OIDCProviderURL  string `gorm:"size:1024;default:''" json:"oidc_provider_url"`
	OIDCClientID     string `gorm:"size:255;default:''" json:"oidc_client_id"`
	OIDCClientSecret string `gorm:"size:255;default:''" json:"-"`

	// Relevance preferences
	IncludeKeywords  StringSlice `gorm:"type:text;serializer:json" json:"include_keywords"`
	ExcludeKeywords  StringSlice `gorm:"type:text;serializer:json" json:"exclude_keywords"`
	LocationKeywords StringSlice `gorm:"type:text;serializer:json" json:"location_keywords"`
	WorkTypes        StringSlice `gorm:"type:text;serializer:json" json:"work_types"`
	EmploymentType   string      `gorm:"size:50;default:''" json:"employment_type"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Config) TableName() string { return "config" }