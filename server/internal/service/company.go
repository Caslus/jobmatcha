package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
)

const companyFreshnessWindow = 30 * 24 * time.Hour

var ErrCompanyNotFound = repository.ErrCompanyNotFound

// AdapterAvailability is the narrow scanner boundary needed for company
// management. It keeps provider registration authoritative to the scanner.
type AdapterAvailability interface {
	SupportsAdapter(string) bool
}

type CompanyService struct {
	repo     *repository.CompanyRepo
	adapters AdapterAvailability
	now      func() time.Time
}

func NewCompanyService(repo *repository.CompanyRepo, adapters AdapterAvailability) *CompanyService {
	return &CompanyService{repo: repo, adapters: adapters, now: time.Now}
}

func (s *CompanyService) List() ([]model.CompanyListItem, error) {
	companies, err := s.repo.ListAllWithRoleCounts()
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	items := make([]model.CompanyListItem, len(companies))
	for i := range companies {
		items[i] = s.toListItem(&companies[i].Company, companies[i].RoleCount)
	}
	return items, nil
}

func (s *CompanyService) UpdateActive(id uint, active bool) (*model.CompanyListItem, error) {
	if err := s.repo.UpdateActive(id, active); err != nil {
		return nil, err
	}
	company, err := s.repo.GetWithRoleCount(id)
	if err != nil {
		return nil, fmt.Errorf("get updated company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("%w: %d", ErrCompanyNotFound, id)
	}
	item := s.toListItem(&company.Company, company.RoleCount)
	return &item, nil
}

func (s *CompanyService) UpdateActiveBulk(ids []uint, active bool) error {
	return s.repo.UpdateActiveBulk(ids, active)
}

func (s *CompanyService) toListItem(company *model.Company, roleCount int64) model.CompanyListItem {
	supported := s.adapters.SupportsAdapter(company.ATSType)
	return model.CompanyListItem{
		ID:                     company.ID,
		Name:                   company.Name,
		Location:               company.Location,
		ATSType:                company.ATSType,
		Active:                 company.Active,
		RoleCount:              roleCount,
		AdapterStatus:          adapterStatus(company, supported),
		FreshnessStatus:        freshnessStatus(company, supported, s.now()),
		LastScanAttemptAt:      company.LastScanAttemptAt,
		LastSuccessfulScanAt:   company.LastSuccessfulScanAt,
		LastScanFailureDetail:  safeFailureDetail(company.LastScanFailureDetail),
		LastNewRoleDiscoveryAt: company.LastNewRoleDiscoveryAt,
	}
}

func adapterStatus(company *model.Company, supported bool) string {
	if !supported {
		return "unsupported"
	}
	if company.LastScanFailureDetail != nil {
		return "failing"
	}
	if company.LastScanAttemptAt == nil {
		return "unknown"
	}
	return "healthy"
}

func freshnessStatus(company *model.Company, supported bool, now time.Time) string {
	if !company.Active || !supported {
		return "not_applicable"
	}
	if company.LastNewRoleDiscoveryAt == nil {
		return "no_activity_yet"
	}
	if now.Sub(*company.LastNewRoleDiscoveryAt) > companyFreshnessWindow {
		return "stale"
	}
	return "fresh"
}

func safeFailureDetail(detail *string) *string {
	if detail == nil {
		return nil
	}
	value := strings.TrimSpace(*detail)
	if value == "" {
		return nil
	}
	if len(value) > 500 {
		value = value[:500]
	}
	return &value
}
