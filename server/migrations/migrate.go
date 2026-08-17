package migrations

import (
	"github.com/caslus/jobmatcha/internal/model"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.Company{},
		&model.Role{},
		&model.Config{},
		&model.Session{},
	)
}