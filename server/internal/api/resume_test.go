package api

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/caslus/jobmatcha/internal/ai"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type resumeHandlerAI struct {
	parseResult *ai.ParseResumeResult
	tailored    *model.ResumeDocument
}

func (f *resumeHandlerAI) ValidateKey(context.Context, string) (bool, int, error) {
	return true, 1, nil
}

func (f *resumeHandlerAI) ParseResume(context.Context, string, string) (*ai.ParseResumeResult, error) {
	return f.parseResult, nil
}

func (f *resumeHandlerAI) TailorResume(context.Context, string, model.ResumeDocument, string, string, string, string) (*model.ResumeDocument, error) {
	return f.tailored, nil
}

func setupResumeRouter(t *testing.T, db *gorm.DB, client service.AIClient) (*gin.Engine, *repository.Repositories) {
	t.Helper()
	repos := repository.NewRepositories(db)
	handler := NewResumeHandler(repos, client)
	router := gin.New()
	router.POST("/api/ai/parse-resume", handler.ParseUpload)
	router.POST("/api/roles/:id/tailor", handler.Tailor)
	router.GET("/api/roles/:id/tailored-resume", handler.GetTailored)
	return router, repos
}

func resumeUpload(t *testing.T, router *gin.Engine, filename, content string) *httptest.ResponseRecorder {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatalf("create form file: %v", err)
	}
	if _, err := part.Write([]byte(content)); err != nil {
		t.Fatalf("write form file: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/ai/parse-resume", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestResumeHandlerParseUploadValidationAndPersistence(t *testing.T) {
	db := setupTestDB(t)
	client := &resumeHandlerAI{parseResult: &ai.ParseResumeResult{
		Name: "Ada Lovelace", Email: "ada@example.test", Location: "London",
		Document: model.ResumeDocument{Header: model.ResumeHeader{Name: "Ada Lovelace"}},
	}}
	router, repos := setupResumeRouter(t, db, client)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/ai/parse-resume", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing file status = %d: %s", w.Code, w.Body.String())
	}

	w = resumeUpload(t, router, "resume.docx", "not supported")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("unsupported upload status = %d: %s", w.Code, w.Body.String())
	}

	if err := repos.Config.Create(context.Background(), &model.Config{ID: 1, AIApiKey: "test-key"}); err != nil {
		t.Fatalf("create AI configuration: %v", err)
	}
	w = resumeUpload(t, router, "resume.txt", "Ada's resume")
	if w.Code != http.StatusOK {
		t.Fatalf("parse upload status = %d: %s", w.Code, w.Body.String())
	}
	var response model.ParseResumeResponse
	parseJSON(t, w.Body.String(), &response)
	if response.Name != "Ada Lovelace" || response.Resume.ID == 0 || response.Resume.Filename != "resume.txt" {
		t.Fatalf("parse response = %#v", response)
	}
	persisted, err := repos.Resume.GetLatest()
	if err != nil || persisted == nil || persisted.Document.Header.Name != "Ada Lovelace" {
		t.Fatalf("persisted resume = %#v, err = %v", persisted, err)
	}
}

func TestResumeHandlerTailorAndRetrieveWorkflow(t *testing.T) {
	db := setupTestDB(t)
	tailoredDocument := &model.ResumeDocument{Header: model.ResumeHeader{Name: "Ada Lovelace"}, Summary: "Tailored for platform engineering"}
	router, repos := setupResumeRouter(t, db, &resumeHandlerAI{tailored: tailoredDocument})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/roles/not-a-number/tailor", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid role status = %d: %s", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/roles/1/tailor", nil))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing resume status = %d: %s", w.Code, w.Body.String())
	}

	if err := repos.Config.Create(context.Background(), &model.Config{ID: 1, AIApiKey: "test-key"}); err != nil {
		t.Fatalf("create AI configuration: %v", err)
	}
	company := &model.Company{Name: "Analytical Engines", CareersURL: "https://example.test/careers"}
	if err := db.Create(company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	role := &model.Role{CompanyID: company.ID, URLHash: "api-resume-role", URL: "https://example.test/jobs/1", Title: "Platform Engineer"}
	if err := repos.Role.Create(role); err != nil {
		t.Fatalf("create role: %v", err)
	}
	resume := &model.Resume{Filename: "resume.txt", Content: "Ada's resume", Document: model.ResumeDocument{Header: model.ResumeHeader{Name: "Ada"}}}
	if err := repos.Resume.Create(resume); err != nil {
		t.Fatalf("create resume: %v", err)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/roles/1/tailor", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("tailor status = %d: %s", w.Code, w.Body.String())
	}
	var tailored model.TailoredResumeResponse
	parseJSON(t, w.Body.String(), &tailored)
	if tailored.ID == 0 || tailored.RoleID != role.ID || tailored.Document.Summary != tailoredDocument.Summary {
		t.Fatalf("tailored response = %#v", tailored)
	}

	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/roles/1/tailored-resume", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("get tailored status = %d: %s", w.Code, w.Body.String())
	}
	var retrieved model.TailoredResumeResponse
	parseJSON(t, w.Body.String(), &retrieved)
	if retrieved.ID != tailored.ID || retrieved.Document.Summary != tailoredDocument.Summary {
		t.Fatalf("retrieved response = %#v", retrieved)
	}
}
