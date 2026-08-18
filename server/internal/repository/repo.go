package repository

import "gorm.io/gorm"

type Repositories struct {
	Role    *RoleRepo
	Company *CompanyRepo
	Config  *ConfigRepo
	ScanJob *ScanJobRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Role:    NewRoleRepo(db),
		Company: NewCompanyRepo(db),
		Config:  NewConfigRepo(db),
		ScanJob: NewScanJobRepo(db),
	}
}