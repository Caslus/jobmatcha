package model

import "time"

type Company struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;not null;size:255" json:"name"`
	CareersURL  string    `gorm:"not null;size:1024" json:"careers_url"`
	ATSType     string    `gorm:"size:50" json:"ats_type"`
	ATSSlug     string    `gorm:"size:255" json:"ats_slug"`
	Region      string    `gorm:"size:10;not null;default:JP" json:"region"`
	Location    string    `gorm:"size:255" json:"location"`
	Active      bool      `gorm:"not null;default:true" json:"active"`
	LastScannedAt *time.Time `json:"last_scanned_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Company) TableName() string { return "companies" }