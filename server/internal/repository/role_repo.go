package repository

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

type RoleRepo struct{ db *gorm.DB }

func NewRoleRepo(db *gorm.DB) *RoleRepo { return &RoleRepo{db: db} }

func (r *RoleRepo) List(page, perPage int) ([]model.Role, int64, error) {
	var total int64
	r.db.Model(&model.Role{}).Count(&total)

	var roles []model.Role
	offset := (page - 1) * perPage
	if err := r.db.Preload("Company").Order("created_at DESC").Offset(offset).Limit(perPage).Find(&roles).Error; err != nil {
		return nil, 0, err
	}
	return roles, total, nil
}

func (r *RoleRepo) ListAll() ([]model.Role, error) {
	var roles []model.Role
	if err := r.db.Preload("Company").Order("created_at DESC").Find(&roles).Error; err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *RoleRepo) GetByID(id uint) (*model.Role, error) {
	var role model.Role
	result := r.db.Preload("Company").Where("id = ?", id).Find(&role)
	if result.RowsAffected == 0 {
		return nil, nil
	}
	return &role, nil
}

func (r *RoleRepo) Patch(id uint, updates map[string]interface{}) error {
	return r.db.Model(&model.Role{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RoleRepo) Create(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *RoleRepo) BulkCreate(roles []model.Role) error {
	if len(roles) == 0 {
		return nil
	}
	return r.db.CreateInBatches(roles, 100).Error
}