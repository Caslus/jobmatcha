package model

import (
	"strings"
	"time"
)

// Resume stores the extracted, searchable text from an uploaded resume.
// The application is currently single-user, so the most recently uploaded
// resume is used when tailoring a role.
type Resume struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	Filename         string         `gorm:"not null;size:512" json:"filename"`
	MediaType        string         `gorm:"size:255;default:''" json:"media_type"`
	Content          string         `gorm:"type:text;not null" json:"-"`
	FormattedContent string         `gorm:"type:text;not null;default:''" json:"-"`
	Document         ResumeDocument `gorm:"type:text;serializer:json;not null" json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (Resume) TableName() string { return "resumes" }

// ResumeInfoResponse intentionally excludes the original resume text.
type ResumeInfoResponse struct {
	ID        uint      `json:"id"`
	Filename  string    `json:"filename"`
	MediaType string    `json:"media_type"`
	CreatedAt time.Time `json:"created_at"`
}

type ResumeHeader struct {
	Name    string   `json:"name"`
	Contact []string `json:"contact"`
}

type ResumeEntry struct {
	Title        string   `json:"title"`
	Organization string   `json:"organization"`
	Location     string   `json:"location"`
	DateRange    string   `json:"date_range"`
	Highlights   []string `json:"highlights"`
}

type ResumeSection struct {
	Heading string        `json:"heading"`
	Kind    string        `json:"kind"`
	Entries []ResumeEntry `json:"entries"`
	Items   []string      `json:"items"`
}

// ResumeDocument is a layout-ready representation of a resume. Content is
// retained only for backward compatibility with previously generated records.
type ResumeDocument struct {
	Content  string          `json:"content,omitempty"`
	Header   ResumeHeader    `json:"header"`
	Summary  string          `json:"summary"`
	Sections []ResumeSection `json:"sections"`
}

func (d ResumeDocument) IsStructured() bool {
	return d.Header.Name != "" || len(d.Sections) > 0
}

// NeedsLayoutRefinement identifies a first-generation document where a long
// skills list was represented as one item per skill. Re-parsing the original
// resume turns that into compact, labelled rows before it is tailored.
func (d ResumeDocument) NeedsLayoutRefinement() bool {
	for _, section := range d.Sections {
		if section.Kind != "list" || len(section.Items) < 10 {
			continue
		}
		groupedRows := 0
		for _, item := range section.Items {
			if strings.Contains(item, ":") {
				groupedRows++
			}
		}
		if groupedRows == 0 {
			return true
		}
	}
	return false
}

// TailoredResume persists a generated document for a role and source resume.
// Regenerating replaces the document for that pair while preserving its ID.
type TailoredResume struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	ResumeID  uint           `gorm:"not null;uniqueIndex:idx_tailored_resume_role" json:"resume_id"`
	RoleID    uint           `gorm:"not null;uniqueIndex:idx_tailored_resume_role" json:"role_id"`
	Document  ResumeDocument `gorm:"type:text;serializer:json;not null" json:"document"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

func (TailoredResume) TableName() string { return "tailored_resumes" }

type TailoredResumeResponse struct {
	ID        uint           `json:"id"`
	ResumeID  uint           `json:"resume_id"`
	RoleID    uint           `json:"role_id"`
	Document  ResumeDocument `json:"document"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}
