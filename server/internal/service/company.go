package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
)

const companyFreshnessWindow = 30 * 24 * time.Hour

var ErrCompanyNotFound = repository.ErrCompanyNotFound
var ErrCareerBoardNotFound = repository.ErrCareerBoardNotFound

type CareerBoardNormalizer interface {
	NormalizeCareerBoard(context.Context, model.BoardIdentity) (model.BoardIdentity, error)
}

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

func (s *CompanyService) UpdateDetails(id uint, name, location string) (*model.CompanyListItem, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("company name is required")
	}
	if err := s.repo.UpdateDetails(id, name, strings.TrimSpace(location)); err != nil {
		return nil, err
	}
	return s.companyItem(id)
}

func (s *CompanyService) Delete(id uint) error {
	return s.repo.Delete(id)
}

func (s *CompanyService) UpdateBoardActive(companyID, boardID uint, active bool) (*model.CompanyListItem, error) {
	if err := s.repo.UpdateBoardActive(companyID, boardID, active); err != nil {
		return nil, err
	}
	company, err := s.repo.GetWithRoleCount(companyID)
	if err != nil {
		return nil, fmt.Errorf("get updated company: %w", err)
	}
	if company == nil {
		return nil, fmt.Errorf("%w: %d", ErrCompanyNotFound, companyID)
	}
	item := s.toListItem(&company.Company, company.RoleCount)
	return &item, nil
}

func (s *CompanyService) CreateBoard(ctx context.Context, companyID uint, input model.CareerBoardUpsertRequest) (*model.CompanyListItem, error) {
	identity, err := s.normalizeBoard(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateBoard(companyID, &model.CareerBoard{Provider: identity.Provider, BoardIdentifier: identity.BoardIdentifier, CanonicalURL: identity.CanonicalURL, Active: true}); err != nil {
		return nil, err
	}
	return s.companyItem(companyID)
}

func (s *CompanyService) UpdateBoardDetails(ctx context.Context, companyID, boardID uint, input model.CareerBoardUpsertRequest) (*model.CompanyListItem, error) {
	identity, err := s.normalizeBoard(ctx, input)
	if err != nil {
		return nil, err
	}
	if err := s.repo.UpdateBoardDetails(companyID, boardID, identity); err != nil {
		return nil, err
	}
	return s.companyItem(companyID)
}

func (s *CompanyService) DeleteBoard(companyID, boardID uint) (*model.CompanyListItem, error) {
	if err := s.repo.DeleteBoard(companyID, boardID); err != nil {
		return nil, err
	}
	return s.companyItem(companyID)
}

func (s *CompanyService) RegisterBoards(selections []model.CareerBoardRegistration) error {
	if len(selections) == 0 {
		return nil
	}
	for _, selection := range selections {
		if strings.TrimSpace(selection.CompanyName) == "" || strings.TrimSpace(selection.CareersURL) == "" || strings.TrimSpace(selection.Provider) == "" || strings.TrimSpace(selection.BoardIdentifier) == "" || strings.TrimSpace(selection.CanonicalURL) == "" {
			return fmt.Errorf("invalid career board selection")
		}
	}
	return s.repo.RegisterBoards(selections, s.adapters.SupportsAdapter)
}

func (s *CompanyService) SuggestedEmployerName(candidates []model.CareerBoardDiscoveryCandidate) (string, error) {
	for _, candidate := range candidates {
		name, err := s.repo.FindCompanyNameByBoard(candidate.Provider, candidate.BoardIdentifier)
		if err != nil {
			return "", fmt.Errorf("find company for discovered board: %w", err)
		}
		if name != "" {
			return name, nil
		}
	}
	return "", nil
}

func (s *CompanyService) normalizeBoard(ctx context.Context, input model.CareerBoardUpsertRequest) (model.BoardIdentity, error) {
	identity := model.BoardIdentity{Provider: strings.ToLower(strings.TrimSpace(input.Provider)), BoardIdentifier: strings.ToLower(strings.TrimSpace(input.BoardIdentifier)), CanonicalURL: strings.TrimSpace(input.CanonicalURL)}
	if identity.Provider == "" || identity.BoardIdentifier == "" || identity.CanonicalURL == "" || !s.adapters.SupportsAdapter(identity.Provider) {
		return model.BoardIdentity{}, fmt.Errorf("unsupported or incomplete career board")
	}
	normalizer, ok := s.adapters.(CareerBoardNormalizer)
	if !ok {
		return model.BoardIdentity{}, fmt.Errorf("career board validation is unavailable")
	}
	return normalizer.NormalizeCareerBoard(ctx, identity)
}

func (s *CompanyService) companyItem(id uint) (*model.CompanyListItem, error) {
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

func (s *CompanyService) toListItem(company *model.Company, roleCount int64) model.CompanyListItem {
	boards := boardListItems(company.CareerBoards, s.adapters, s.now())
	return model.CompanyListItem{
		ID:                     company.ID,
		Name:                   company.Name,
		Location:               company.Location,
		Active:                 company.Active,
		BoardCount:             len(boards),
		RoleCount:              roleCount,
		FreshnessStatus:        aggregateFreshnessStatus(boards),
		LastScanAttemptAt:      latestBoardTime(boards, func(board model.CareerBoardListItem) *time.Time { return board.LastScanAttemptAt }),
		LastNewRoleDiscoveryAt: latestBoardTime(boards, func(board model.CareerBoardListItem) *time.Time { return board.LastNewRoleDiscoveryAt }),
		CareerBoards:           boards,
	}
}

func latestBoardTime(boards []model.CareerBoardListItem, selectTime func(model.CareerBoardListItem) *time.Time) *time.Time {
	var latest *time.Time
	for _, board := range boards {
		at := selectTime(board)
		if at != nil && (latest == nil || at.After(*latest)) {
			value := *at
			latest = &value
		}
	}
	return latest
}

func aggregateFreshnessStatus(boards []model.CareerBoardListItem) string {
	if len(boards) == 0 {
		return "unknown"
	}
	hasFresh, hasNoActivity, hasApplicable := false, false, false
	for _, board := range boards {
		if board.AdapterStatus == "failing" {
			return "failing"
		}
		switch board.FreshnessStatus {
		case "stale":
			return "stale"
		case "fresh":
			hasFresh, hasApplicable = true, true
		case "no_activity_yet":
			hasNoActivity, hasApplicable = true, true
		}
	}
	if hasFresh {
		return "fresh"
	}
	if hasNoActivity || hasApplicable {
		return "no_activity_yet"
	}
	return "not_applicable"
}

func boardListItems(boards []model.CareerBoard, adapters AdapterAvailability, now time.Time) []model.CareerBoardListItem {
	items := make([]model.CareerBoardListItem, len(boards))
	for i, board := range boards {
		supported := adapters.SupportsAdapter(board.Provider)
		items[i] = model.CareerBoardListItem{ID: board.ID, Provider: board.Provider, BoardIdentifier: board.BoardIdentifier, CanonicalURL: board.CanonicalURL, Active: board.Active, AdapterStatus: boardAdapterStatus(&board, supported), FreshnessStatus: boardFreshnessStatus(&board, supported, now), LastScanAttemptAt: board.LastScanAttemptAt, LastSuccessfulScanAt: board.LastSuccessfulScanAt, LastScanFailureDetail: safeFailureDetail(board.LastScanFailureDetail), LastNewRoleDiscoveryAt: board.LastNewRoleDiscoveryAt}
	}
	return items
}

func boardAdapterStatus(board *model.CareerBoard, supported bool) string {
	if !supported {
		return "unsupported"
	}
	if board.LastScanFailureDetail != nil {
		return "failing"
	}
	if board.LastScanAttemptAt == nil {
		return "unknown"
	}
	return "healthy"
}

func boardFreshnessStatus(board *model.CareerBoard, supported bool, now time.Time) string {
	if !board.Active || !supported {
		return "not_applicable"
	}
	if board.LastNewRoleDiscoveryAt == nil {
		return "no_activity_yet"
	}
	if now.Sub(*board.LastNewRoleDiscoveryAt) > companyFreshnessWindow {
		return "stale"
	}
	return "fresh"
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
