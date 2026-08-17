package repository

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db: db} }

// Get returns the single config row, creating it with defaults if it doesn't exist.
func (r *ConfigRepo) Get() (*model.Config, error) {
	var cfg model.Config
	result := r.db.Where("id = 1").Find(&cfg)
	if result.RowsAffected == 0 {
		cfg = model.Config{
			IncludeKeywords:  []string{},
			ExcludeKeywords:  []string{},
			LocationKeywords: []string{},
			WorkTypes:        []string{},
		}
		if err := r.db.Create(&cfg).Error; err != nil {
			return nil, err
		}
		// Re-fetch to get the auto-created fields
		r.db.Where("id = 1").Find(&cfg)
	}
	return &cfg, nil
}

func (r *ConfigRepo) Update(cfg *model.Config) error {
	return r.db.Model(&model.Config{}).Where("id = 1").Updates(cfg).Error
}