package repository

import (
	"github.com/caslus/jobmatcha/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const defaultBcryptCost = 12

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db: db} }

// Get returns the single config row, creating it with defaults if it doesn't exist.
func (r *ConfigRepo) Get() (*model.Config, error) {
	var cfg model.Config
	result := r.db.Where("id = 1").Find(&cfg)
	if result.RowsAffected == 0 {
		// First run — create config with default password
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), defaultBcryptCost)
		if err != nil {
			return nil, err
		}
		cfg = model.Config{
			PasswordHash:     string(hash),
			SetupComplete:    false,
			IncludeKeywords:  []string{},
			ExcludeKeywords:  []string{},
			LocationKeywords: []string{},
			WorkTypes:        []string{},
			MaxDaysOld:       0,
		}
		if err := r.db.Create(&cfg).Error; err != nil {
			return nil, err
		}
		r.db.Where("id = 1").Find(&cfg)
	} else if cfg.PasswordHash == "" {
		// Existing row but no password (upgraded from old version) — backfill default
		hash, err := bcrypt.GenerateFromPassword([]byte("admin"), defaultBcryptCost)
		if err != nil {
			return nil, err
		}
		cfg.PasswordHash = string(hash)
		cfg.SetupComplete = false
		if err := r.db.Model(&model.Config{}).Where("id = 1").Updates(map[string]interface{}{
			"password_hash":  cfg.PasswordHash,
			"setup_complete": false,
		}).Error; err != nil {
			return nil, err
		}
	}
	return &cfg, nil
}

func (r *ConfigRepo) UpdateMap(updates map[string]interface{}) error {
	return r.db.Model(&model.Config{}).Where("id = 1").Updates(updates).Error
}