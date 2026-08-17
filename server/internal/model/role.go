package model

import "time"

type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CompanyID   uint      `gorm:"not null;index" json:"company_id"`
	URLHash     string    `gorm:"uniqueIndex:idx_company_url;not null;size:64" json:"-"`
	URL         string    `gorm:"not null;size:1024" json:"url"`
	Title       string    `gorm:"not null;size:512" json:"title"`
	Department  string    `gorm:"size:255" json:"department"`
	Location    string    `gorm:"size:255" json:"location"`
	Description string    `gorm:"type:text" json:"description"`
	PostedAt    *time.Time `json:"posted_at"`
	Status      string    `gorm:"size:20;not null;default:seen" json:"status"`
	IsHidden    bool      `gorm:"not null;default:false" json:"is_hidden"`
	IsInterested bool     `gorm:"not null;default:false" json:"is_interested"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Role) TableName() string { return "roles" }