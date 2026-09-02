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
	if err != nil {
		return nil, err
	}
	for i := range companies {
		if err := r.db.Where("company_id = ?", companies[i].ID).Order("provider, board_identifier").Find(&companies[i].CareerBoards).Error; err != nil {
			return nil, err
		}
	}
	return companies, nil
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
	if err := r.db.Where("company_id = ?", company.ID).Order("provider, board_identifier").Find(&company.CareerBoards).Error; err != nil {
		return nil, err
	}
	return &company, nil
}

func (r *CompanyRepo) UpdateBoardActive(companyID, boardID uint, active bool) error {
	result := r.db.Model(&model.CareerBoard{}).Where("id = ? AND company_id = ?", boardID, companyID).Update("active", active)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrCareerBoardNotFound, boardID)
	}
	return nil
}

func (r *CompanyRepo) UpdateDetails(id uint, name, location string) error {
	result := r.db.Model(&model.Company{}).Where("id = ?", id).Updates(map[string]interface{}{"name": name, "location": location})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrCompanyNotFound, id)
	}
	return nil
}

// Delete removes a company and its current sources in one transaction. Roles
// deliberately remain as historical records and are hidden by company-aware
// role-list queries.
func (r *CompanyRepo) Delete(id uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		var company model.Company
		if err := tx.First(&company, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %d", ErrCompanyNotFound, id)
			}
			return err
		}
		if err := tx.Where("company_id = ?", id).Delete(&model.CareerBoard{}).Error; err != nil {
			return err
		}
		return tx.Delete(&company).Error
	})
}

func (r *CompanyRepo) CreateBoard(companyID uint, board *model.CareerBoard) error {
	var company model.Company
	if err := r.db.First(&company, companyID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %d", ErrCompanyNotFound, companyID)
		}
		return err
	}
	board.CompanyID = companyID
	return r.db.Create(board).Error
}

func (r *CompanyRepo) UpdateBoardDetails(companyID, boardID uint, board model.BoardIdentity) error {
	result := r.db.Model(&model.CareerBoard{}).Where("id = ? AND company_id = ?", boardID, companyID).Updates(map[string]interface{}{"provider": board.Provider, "board_identifier": board.BoardIdentifier, "canonical_url": board.CanonicalURL})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrCareerBoardNotFound, boardID)
	}
	return nil
}

func (r *CompanyRepo) DeleteBoard(companyID, boardID uint) error {
	result := r.db.Where("id = ? AND company_id = ?", boardID, companyID).Delete(&model.CareerBoard{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("%w: %d", ErrCareerBoardNotFound, boardID)
	}
	return nil
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

func (r *CompanyRepo) FindCompanyNameByBoard(provider, boardIdentifier string) (string, error) {
	var board model.CareerBoard
	result := r.db.Preload("Company").Where("provider = ? AND board_identifier = ?", provider, boardIdentifier).First(&board)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if result.Error != nil {
		return "", result.Error
	}
	return board.Company.Name, nil
}

// RegisterBoards persists only explicit user selections. When one selected
// board already belongs to a company, the other boards discovered from that
// same careers site are attached to that company too.
func (r *CompanyRepo) RegisterBoards(selections []model.CareerBoardRegistration, supports func(string) bool) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		groups := map[string][]model.CareerBoardRegistration{}
		for _, selection := range selections {
			key := "shared"
			if selection.Separate {
				key = selection.Provider + ":" + selection.BoardIdentifier
			}
			groups[key] = append(groups[key], selection)
		}
		for _, group := range groups {
			if err := r.registerBoardGroup(tx, group, supports); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *CompanyRepo) registerBoardGroup(tx *gorm.DB, selections []model.CareerBoardRegistration, supports func(string) bool) error {
	var targetCompanyID uint
	for _, selection := range selections {
		var existing model.CareerBoard
		if err := tx.Where("provider = ? AND board_identifier = ?", selection.Provider, selection.BoardIdentifier).First(&existing).Error; err == nil {
			if targetCompanyID == 0 {
				targetCompanyID = existing.CompanyID
			} else if targetCompanyID != existing.CompanyID {
				targetCompanyID = 0
				break
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	if targetCompanyID == 0 && len(selections) > 0 && !selections[0].Separate {
		var company model.Company
		if err := tx.Where("careers_url = ?", selections[0].CareersURL).First(&company).Error; err == nil {
			targetCompanyID = company.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}

	for _, selection := range selections {
		var existing model.CareerBoard
		if err := tx.Where("provider = ? AND board_identifier = ?", selection.Provider, selection.BoardIdentifier).First(&existing).Error; err == nil {
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		region := selection.Region
		if region == "" {
			region = "JP"
		}
		company := model.Company{ID: targetCompanyID}
		if targetCompanyID == 0 {
			company = model.Company{Name: selection.CompanyName, CareersURL: selection.CareersURL, Location: selection.Location, Region: region, Active: true}
			if err := tx.Where("name = ?", selection.CompanyName).FirstOrCreate(&company).Error; err != nil {
				return err
			}
		}
		board := model.CareerBoard{CompanyID: company.ID, Provider: selection.Provider, BoardIdentifier: selection.BoardIdentifier, CanonicalURL: selection.CanonicalURL, Active: supports(selection.Provider)}
		if err := tx.Create(&board).Error; err != nil {
			return err
		}
	}
	return nil
}
