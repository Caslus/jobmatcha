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
	err := r.db.First(&cfg, 1).Error
	if err == gorm.ErrRecordNotFound {
		cfg = model.Config{
			IncludeKeywords:  []string{},
			ExcludeKeywords:  []string{},
			LocationKeywords: []string{},
			WorkTypes:        []string{},
		}
		if err := r.db.Create(&cfg).Error; err != nil {
			return nil, err
		}
		return &cfg, nil
	}
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (r *ConfigRepo) Update(cfg *model.Config) error {
	return r.db.Model(&model.Config{}).Where("id = 1").Updates(cfg).Error
}