package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

var ErrCompanyNotFound = errors.New("company not found")

type CompanyRepo struct{ db *gorm.DB }

func NewCompanyRepo(db *gorm.DB) *CompanyRepo { return &CompanyRepo{db: db} }

type CompanyWithRoleCount struct {
	model.Company
	RoleCount int64 `gorm:"column:role_count"`
}

func (r *CompanyRepo) ListActive() ([]model.Company, error) {
	var companies []model.Company
	if err := r.db.Where("active = ?", true).Order("name ASC").Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *CompanyRepo) ListAll() ([]model.Company, error) {
	var companies []model.Company
	if err := r.db.Order("name ASC").Find(&companies).Error; err != nil {
		return nil, err
	}
	return companies, nil
}

func (r *CompanyRepo) ListAllWithRoleCounts() ([]CompanyWithRoleCount, error) {
	var companies []CompanyWithRoleCount
	err := r.db.Model(&model.Company{}).
		Select("companies.*, COUNT(roles.id) AS role_count").
		Joins("LEFT JOIN roles ON roles.company_id = companies.id").
		Group("companies.id").
		Order("companies.name ASC").
		Scan(&companies).Error
	return companies, err
}

func (r *CompanyRepo) GetWithRoleCount(id uint) (*CompanyWithRoleCount, error) {
	var company CompanyWithRoleCount
	result := r.db.Model(&model.Company{}).
		Select("companies.*, COUNT(roles.id) AS role_count").
		Joins("LEFT JOIN roles ON roles.company_id = companies.id").
		Where("companies.id = ?", id).
		Group("companies.id").
		Scan(&company)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &company, nil
}

func (r *CompanyRepo) GetByID(id uint) (*model.Company, error) {
	var company model.Company
	result := r.db.Where("id = ?", id).Find(&company)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &company, nil
}

func (r *CompanyRepo) RecordScanAttempt(id uint, at time.Time) error {
	return r.db.Model(&model.Company{}).Where("id = ?", id).Update("last_scan_attempt_at", at).Error
}

func (r *CompanyRepo) RecordScanSuccess(id uint, at time.Time) error {
	return r.db.Model(&model.Company{}).Where("id = ?", id).Updates(map[string]interface{}{
		"last_scanned_at":          at,
		"last_successful_scan_at":  at,
		"last_scan_failure_detail": nil,
	}).Error
}

func (r *CompanyRepo) RecordScanFailure(id uint, detail string) error {
	return r.db.Model(&model.Company{}).Where("id = ?", id).Update("last_scan_failure_detail", detail).Error
}

func (r *CompanyRepo) RecordNewRoleDiscovery(id uint, at time.Time) error {
	return r.db.Model(&model.Company{}).Where("id = ?", id).Update("last_new_role_discovery_at", at).Error
}

func (r *CompanyRepo) UpdateActive(id uint, active bool) error {
	result := r.db.Model(&model.Company{}).Where("id = ?", id).Update("active", active)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrCompanyNotFound, id)
	}
	return nil
}

// UpdateActiveBulk updates all requested IDs in one transaction. Unknown IDs
// are rejected so callers never receive a partial state update.
func (r *CompanyRepo) UpdateActiveBulk(ids []uint, active bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&model.Company{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(ids)) {
			return ErrCompanyNotFound
		}
		return tx.Model(&model.Company{}).Where("id IN ?", ids).Update("active", active).Error
	})
}
