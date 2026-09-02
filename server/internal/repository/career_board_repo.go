package repository

import (
	"errors"
	"fmt"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

var ErrCareerBoardNotFound = errors.New("career board not found")

type CareerBoardRepo struct{ db *gorm.DB }

func NewCareerBoardRepo(db *gorm.DB) *CareerBoardRepo { return &CareerBoardRepo{db: db} }

func (r *CareerBoardRepo) Create(board *model.CareerBoard) error {
	return r.db.Create(board).Error
}

func (r *CareerBoardRepo) ListForCompany(companyID uint) ([]model.CareerBoard, error) {
	var boards []model.CareerBoard
	if err := r.db.Where("company_id = ?", companyID).Order("provider, board_identifier").Find(&boards).Error; err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *CareerBoardRepo) ListEnabledWithCompanies() ([]model.CareerBoard, error) {
	var boards []model.CareerBoard
	if err := r.db.Joins("Company").Where("career_boards.active = ? AND Company.active = ?", true, true).Order("career_boards.id").Find(&boards).Error; err != nil {
		return nil, err
	}
	return boards, nil
}

func (r *CareerBoardRepo) UpdateActive(id uint, active bool) error {
	result := r.db.Model(&model.CareerBoard{}).Where("id = ?", id).Update("active", active)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrCareerBoardNotFound, id)
	}
	return nil
}

func (r *CareerBoardRepo) RecordScanAttempt(id uint, at time.Time) error {
	return r.db.Model(&model.CareerBoard{}).Where("id = ?", id).Update("last_scan_attempt_at", at).Error
}

func (r *CareerBoardRepo) RecordScanSuccess(id uint, at time.Time) error {
	return r.db.Model(&model.CareerBoard{}).Where("id = ?", id).Updates(map[string]interface{}{"last_scanned_at": at, "last_successful_scan_at": at, "last_scan_failure_detail": nil}).Error
}

func (r *CareerBoardRepo) RecordScanFailure(id uint, detail string) error {
	return r.db.Model(&model.CareerBoard{}).Where("id = ?", id).Update("last_scan_failure_detail", detail).Error
}

func (r *CareerBoardRepo) RecordNewRoleDiscovery(id uint, at time.Time) error {
	return r.db.Model(&model.CareerBoard{}).Where("id = ?", id).Update("last_new_role_discovery_at", at).Error
}
