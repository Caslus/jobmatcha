package model

import "time"

type Session struct {
	ID        uint      `gorm:"primaryKey" json:"-"`
	Token     string    `gorm:"uniqueIndex;not null;size:128" json:"-"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"-"`
}

func (Session) TableName() string { return "sessions" }