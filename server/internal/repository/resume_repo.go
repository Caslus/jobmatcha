package repository

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

type ResumeRepo struct{ db *gorm.DB }

func NewResumeRepo(db *gorm.DB) *ResumeRepo { return &ResumeRepo{db: db} }

func (r *ResumeRepo) Create(resume *model.Resume) error {
	return r.db.Create(resume).Error
}

func (r *ResumeRepo) UpdateFormattedContent(id uint, content string) error {
	return r.db.Model(&model.Resume{}).Where("id = ?", id).Update("formatted_content", content).Error
}

func (r *ResumeRepo) UpdateDocument(id uint, document model.ResumeDocument) error {
	var resume model.Resume
	if err := r.db.First(&resume, id).Error; err != nil {
		return err
	}
	resume.Document = document
	return r.db.Save(&resume).Error
}

func (r *ResumeRepo) GetLatest() (*model.Resume, error) {
	var resume model.Resume
	result := r.db.Order("created_at DESC").First(&resume)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &resume, nil
}

func (r *ResumeRepo) UpsertTailored(tailored *model.TailoredResume) error {
	var existing model.TailoredResume
	result := r.db.Where("resume_id = ? AND role_id = ?", tailored.ResumeID, tailored.RoleID).First(&existing)
	if result.Error == nil {
		existing.Document = tailored.Document
		if err := r.db.Save(&existing).Error; err != nil {
			return err
		}
		*tailored = existing
		return nil
	}
	if result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}
	return r.db.Create(tailored).Error
}

func (r *ResumeRepo) GetTailored(resumeID, roleID uint) (*model.TailoredResume, error) {
	var tailored model.TailoredResume
	result := r.db.Where("resume_id = ? AND role_id = ?", resumeID, roleID).First(&tailored)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, result.Error
	}
	return &tailored, nil
}
