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
		}
		if err := r.db.Create(&cfg).Error; err != nil {
			return nil, err
		}
		r.db.Where("id = 1").Find(&cfg)
	}
	return &cfg, nil
}

func (r *ConfigRepo) Update(cfg *model.Config) error {
	return r.db.Model(&model.Config{}).Where("id = 1").Updates(cfg).Error
}