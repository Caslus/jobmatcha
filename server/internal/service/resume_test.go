package service

import (
	"context"
	"errors"
	"testing"

	"github.com/caslus/jobmatcha/internal/ai"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/testutil"
)

type fakeResumeAI struct {
	parseResult  *ai.ParseResumeResult
	parseErr     error
	tailored     *model.ResumeDocument
	tailorErr    error
	parseCalls   int
	tailorCalls  int
	lastTailored struct {
		key, title, company, location, description string
	}
}

func (f *fakeResumeAI) ValidateKey(context.Context, string) (bool, int, error) { return true, 1, nil }

func (f *fakeResumeAI) ParseResume(_ context.Context, _, _ string) (*ai.ParseResumeResult, error) {
	f.parseCalls++
	return f.parseResult, f.parseErr
}

func (f *fakeResumeAI) TailorResume(_ context.Context, key string, _ model.ResumeDocument, title, company, location, description string) (*model.ResumeDocument, error) {
	f.tailorCalls++
	f.lastTailored.key = key
	f.lastTailored.title = title
	f.lastTailored.company = company
	f.lastTailored.location = location
	f.lastTailored.description = description
	return f.tailored, f.tailorErr
}

func TestResumeServiceSaveRejectsEmptyContentAndPersistsUpload(t *testing.T) {
	db := testutil.Database(t)
	service := NewResumeServiceWithAI(testutil.Repositories(db), &fakeResumeAI{})

	if _, err := service.Save(context.Background(), "resume.txt", "text/plain", ""); err == nil {
		t.Fatal("Save accepted empty content")
	}
	resume, err := service.Save(context.Background(), "resume.txt", "text/plain", "Ada's resume")
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if resume.ID == 0 || resume.Content != "Ada's resume" || resume.MediaType != "text/plain" {
		t.Fatalf("saved resume = %#v", resume)
	}
}

func TestResumeServiceParseRequiresKeyAndPersistsStructuredDocument(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	resume := &model.Resume{Filename: "resume.txt", Content: "Ada's resume"}
	if err := repos.Resume.Create(resume); err != nil {
		t.Fatalf("create resume: %v", err)
	}
	if err := repos.Config.Create(ctx, &model.Config{ID: 1}); err != nil {
		t.Fatalf("create config without key: %v", err)
	}
	service := NewResumeServiceWithAI(repos, &fakeResumeAI{})
	if _, err := service.Parse(ctx, resume); !errors.Is(err, ErrAIKeyNotConfigured) {
		t.Fatalf("Parse without key error = %v, want ErrAIKeyNotConfigured", err)
	}

	if err := repos.Config.UpdateMap(ctx, map[string]interface{}{"ai_api_key": "key"}); err != nil {
		t.Fatalf("set config key: %v", err)
	}
	client := &fakeResumeAI{parseResult: &ai.ParseResumeResult{Document: model.ResumeDocument{
		Header:   model.ResumeHeader{Name: "Ada Lovelace"},
		Sections: []model.ResumeSection{{Heading: "Experience", Kind: "experience"}},
	}}}
	service = NewResumeServiceWithAI(repos, client)
	result, err := service.Parse(ctx, resume)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if client.parseCalls != 1 || result.Document.Header.Name != "Ada Lovelace" {
		t.Fatalf("parse result = %#v, calls = %d", result, client.parseCalls)
	}
	persisted, err := repos.Resume.GetLatest()
	if err != nil || persisted.Document.Header.Name != "Ada Lovelace" {
		t.Fatalf("persisted resume = %#v, err = %v", persisted, err)
	}
}

func TestResumeServiceTailorStructuresResumeAndUpsertsResult(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	if err := repos.Config.Create(ctx, &model.Config{ID: 1, AIApiKey: "key"}); err != nil {
		t.Fatalf("create config: %v", err)
	}
	company := &model.Company{Name: "Analytical Engines", CareersURL: "https://example.test/careers"}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	role := &model.Role{CompanyID: company.ID, URLHash: "role-1", URL: "https://example.test/jobs/1", Title: "Engineer", Location: "London", Description: "Build engines"}
	if err := repos.Role.Create(role); err != nil {
		t.Fatalf("create role: %v", err)
	}
	resume := &model.Resume{Filename: "resume.txt", Content: "unstructured text"}
	if err := repos.Resume.Create(resume); err != nil {
		t.Fatalf("create resume: %v", err)
	}
	structured := model.ResumeDocument{Header: model.ResumeHeader{Name: "Ada"}}
	tailoredDocument := &model.ResumeDocument{Header: model.ResumeHeader{Name: "Ada Lovelace"}, Summary: "Tailored"}
	client := &fakeResumeAI{parseResult: &ai.ParseResumeResult{Document: structured}, tailored: tailoredDocument}
	service := NewResumeServiceWithAI(repos, client)

	tailored, err := service.Tailor(ctx, role.ID)
	if err != nil {
		t.Fatalf("Tailor: %v", err)
	}
	if tailored.Document.Summary != "Tailored" || client.parseCalls != 1 || client.tailorCalls != 1 {
		t.Fatalf("tailored = %#v, parse calls = %d, tailor calls = %d", tailored, client.parseCalls, client.tailorCalls)
	}
	if client.lastTailored.key != "key" || client.lastTailored.title != role.Title || client.lastTailored.company != company.Name {
		t.Fatalf("tailor request = %#v", client.lastTailored)
	}

	client.tailored = &model.ResumeDocument{Header: model.ResumeHeader{Name: "Ada"}, Summary: "Updated"}
	updated, err := service.Tailor(ctx, role.ID)
	if err != nil {
		t.Fatalf("second Tailor: %v", err)
	}
	if updated.ID != tailored.ID || updated.Document.Summary != "Updated" || client.parseCalls != 1 {
		t.Fatalf("updated tailored = %#v, parse calls = %d", updated, client.parseCalls)
	}
	loaded, err := service.GetTailored(ctx, role.ID)
	if err != nil || loaded == nil || loaded.ID != tailored.ID {
		t.Fatalf("GetTailored = %#v, err = %v", loaded, err)
	}
}

func TestResumeServiceTailorReportsMissingResumeAndRole(t *testing.T) {
	ctx := context.Background()
	db := testutil.Database(t)
	repos := testutil.Repositories(db)
	service := NewResumeServiceWithAI(repos, &fakeResumeAI{})
	if _, err := service.Tailor(ctx, 1); !errors.Is(err, ErrNoResume) {
		t.Fatalf("Tailor without resume error = %v, want ErrNoResume", err)
	}
	if err := repos.Resume.Create(&model.Resume{Filename: "resume.txt", Content: "text"}); err != nil {
		t.Fatalf("create resume: %v", err)
	}
	if _, err := service.Tailor(ctx, 99); !errors.Is(err, ErrRoleNotFound) {
		t.Fatalf("Tailor without role error = %v, want ErrRoleNotFound", err)
	}
}
