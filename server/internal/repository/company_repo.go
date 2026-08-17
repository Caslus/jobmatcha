package repository

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

type CompanyRepo struct{ db *gorm.DB }

func NewCompanyRepo(db *gorm.DB) *CompanyRepo { return &CompanyRepo{db: db} }

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

func (r *CompanyRepo) GetByID(id uint) (*model.Company, error) {
	var company model.Company
	if err := r.db.First(&company, id).Error; err != nil {
		return nil, err
	}
	return &company, nil
}