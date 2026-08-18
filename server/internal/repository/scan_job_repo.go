package repository

import (
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

type ScanJobRepo struct{ db *gorm.DB }

func NewScanJobRepo(db *gorm.DB) *ScanJobRepo { return &ScanJobRepo{db: db} }

func (r *ScanJobRepo) Create(job *model.ScanJob) error {
	return r.db.Create(job).Error
}

func (r *ScanJobRepo) GetByID(id uint) (*model.ScanJob, error) {
	var job model.ScanJob
	result := r.db.Where("id = ?", id).Find(&job)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &job, nil
}

func (r *ScanJobRepo) GetLatest() (*model.ScanJob, error) {
	var job model.ScanJob
	result := r.db.Order("id DESC").Find(&job)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &job, nil
}

// UpdateStatus transitions a job's lifecycle fields.
func (r *ScanJobRepo) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.ScanJob{}).Where("id = ?", id).Update("status", status).Error
}

// Complete stores results and marks the job finished.
func (r *ScanJobRepo) Complete(id uint, resultsJSON string, durationMS int64) error {
	return r.db.Model(&model.ScanJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "completed",
		"results":      resultsJSON,
		"error":        "",
		"duration_ms":  durationMS,
		"completed_at": time.Now(),
	}).Error
}

// Fail marks the job as failed with an error message.
func (r *ScanJobRepo) Fail(id uint, errMsg string) error {
	return r.db.Model(&model.ScanJob{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":       "failed",
		"error":        errMsg,
		"completed_at": time.Now(),
	}).Error
}
