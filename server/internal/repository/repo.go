package repository

import "gorm.io/gorm"

type Repositories struct {
	Role    *RoleRepo
	Company *CompanyRepo
	Config  *ConfigRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Role:    NewRoleRepo(db),
		Company: NewCompanyRepo(db),
		Config:  NewConfigRepo(db),
	}
}