package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/caslus/jobmatcha/internal/ai"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
)

var (
	ErrNoResume           = errors.New("no uploaded resume")
	ErrRoleNotFound       = errors.New("role not found")
	ErrAIKeyNotConfigured = errors.New("ai api key is not configured")
)

type ResumeService struct {
	resumeRepo *repository.ResumeRepo
	roleRepo   *repository.RoleRepo
	cfgRepo    *repository.ConfigRepo
}

func NewResumeService(repos *repository.Repositories) *ResumeService {
	return &ResumeService{
		resumeRepo: repos.Resume,
		roleRepo:   repos.Role,
		cfgRepo:    repos.Config,
	}
}

func (s *ResumeService) Save(ctx context.Context, filename, mediaType, content string) (*model.Resume, error) {
	if content == "" {
		return nil, fmt.Errorf("saving resume: content is empty")
	}
	resume := &model.Resume{Filename: filename, MediaType: mediaType, Content: content}
	if err := s.resumeRepo.Create(resume); err != nil {
		return nil, fmt.Errorf("saving uploaded resume: %w", err)
	}
	return resume, nil
}

func (s *ResumeService) Parse(ctx context.Context, resume *model.Resume) (*ai.ParseResumeResult, error) {
	cfg, err := s.cfgRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("loading AI configuration: %w", err)
	}
	if cfg.AIApiKey == "" {
		return nil, ErrAIKeyNotConfigured
	}

	result, err := ai.ParseResume(ctx, cfg.AIApiKey, resume.Content)
	if err != nil {
		return nil, fmt.Errorf("parsing resume %d: %w", resume.ID, err)
	}
	if err := s.resumeRepo.UpdateDocument(resume.ID, result.Document); err != nil {
		return nil, fmt.Errorf("saving structured resume %d: %w", resume.ID, err)
	}
	resume.Document = result.Document
	return result, nil
}

func (s *ResumeService) Tailor(ctx context.Context, roleID uint) (*model.TailoredResume, error) {
	resume, err := s.resumeRepo.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("loading latest resume: %w", err)
	}
	if resume == nil {
		return nil, ErrNoResume
	}

	role, err := s.roleRepo.GetByID(roleID)
	if err != nil {
		return nil, fmt.Errorf("loading role %d: %w", roleID, err)
	}
	if role == nil {
		return nil, ErrRoleNotFound
	}

	cfg, err := s.cfgRepo.Get()
	if err != nil {
		return nil, fmt.Errorf("loading AI configuration: %w", err)
	}
	if cfg.AIApiKey == "" {
		return nil, ErrAIKeyNotConfigured
	}

	if !resume.Document.IsStructured() || resume.Document.NeedsLayoutRefinement() {
		if _, err := s.Parse(ctx, resume); err != nil {
			return nil, fmt.Errorf("structuring resume before tailoring: %w", err)
		}
	}
	document, err := ai.TailorResume(ctx, cfg.AIApiKey, resume.Document, role.Title, role.Company.Name, role.Location, role.Description)
	if err != nil {
		return nil, fmt.Errorf("tailoring resume for role %d: %w", roleID, err)
	}

	tailored := &model.TailoredResume{ResumeID: resume.ID, RoleID: roleID, Document: *document}
	if err := s.resumeRepo.UpsertTailored(tailored); err != nil {
		return nil, fmt.Errorf("saving tailored resume for role %d: %w", roleID, err)
	}
	saved, err := s.resumeRepo.GetTailored(resume.ID, roleID)
	if err != nil {
		return nil, fmt.Errorf("loading saved tailored resume: %w", err)
	}
	if saved == nil {
		return nil, fmt.Errorf("loading saved tailored resume: record missing")
	}
	return saved, nil
}

func (s *ResumeService) GetTailored(ctx context.Context, roleID uint) (*model.TailoredResume, error) {
	resume, err := s.resumeRepo.GetLatest()
	if err != nil {
		return nil, fmt.Errorf("loading latest resume: %w", err)
	}
	if resume == nil {
		return nil, nil
	}
	tailored, err := s.resumeRepo.GetTailored(resume.ID, roleID)
	if err != nil {
		return nil, fmt.Errorf("loading tailored resume for role %d: %w", roleID, err)
	}
	return tailored, nil
}
