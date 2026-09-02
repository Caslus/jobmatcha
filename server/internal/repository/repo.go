package repository

import "gorm.io/gorm"

type Repositories struct {
	Role        *RoleRepo
	Company     *CompanyRepo
	CareerBoard *CareerBoardRepo
	Config      *ConfigRepo
	Session     *SessionRepo
	ScanJob     *ScanJobRepo
	Resume      *ResumeRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Role:        NewRoleRepo(db),
		Company:     NewCompanyRepo(db),
		CareerBoard: NewCareerBoardRepo(db),
		Config:      NewConfigRepo(db),
		Session:     NewSessionRepo(db),
		ScanJob:     NewScanJobRepo(db),
		Resume:      NewResumeRepo(db),
	}
}
