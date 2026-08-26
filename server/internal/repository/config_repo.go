package repository

import (
	"context"
	"errors"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

var ErrNotFound = errors.New("repository: record not found")

type ConfigRepo struct{ db *gorm.DB }

func NewConfigRepo(db *gorm.DB) *ConfigRepo { return &ConfigRepo{db: db} }

// Get returns the single config row. Bootstrap is owned by the auth service.
func (r *ConfigRepo) Get(ctx context.Context) (*model.Config, error) {
	var cfg model.Config
	result := r.db.WithContext(ctx).Where("id = 1").First(&cfg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, ErrNotFound
	}
	if result.Error != nil {
		return nil, result.Error
	}
	normalizeConfig(&cfg)
	return &cfg, nil
}
func normalizeConfig(cfg *model.Config) {
	if cfg.IncludeKeywords == nil {
		cfg.IncludeKeywords = model.StringSlice{}
	}
	if cfg.ExcludeKeywords == nil {
		cfg.ExcludeKeywords = model.StringSlice{}
	}
	if cfg.LocationKeywords == nil {
		cfg.LocationKeywords = model.StringSlice{}
	}
	if cfg.WorkTypes == nil {
		cfg.WorkTypes = model.StringSlice{}
	}
	if cfg.ScanCronExpr == "" {
		cfg.ScanCronExpr = "0 */6 * * *"
	}
	if cfg.ScanTimezone == "" {
		cfg.ScanTimezone = "UTC"
	}
	if cfg.AIProvider == "" {
		cfg.AIProvider = "openrouter"
	}
}

func (r *ConfigRepo) Create(ctx context.Context, cfg *model.Config) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

func (r *ConfigRepo) UpdateMap(ctx context.Context, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Config{}).Where("id = 1").Updates(updates).Error
}
