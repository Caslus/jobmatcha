package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/ai"
	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/gin-gonic/gin"
)

type validationAI struct {
	valid  bool
	models int
	err    error
}

func newTestHandlerRouter(method, path string, handler gin.HandlerFunc) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Handle(method, path, handler)
	return router
}

func (f validationAI) ValidateKey(context.Context, string) (bool, int, error) {
	return f.valid, f.models, f.err
}

func (validationAI) ParseResume(context.Context, string, string) (*ai.ParseResumeResult, error) {
	return nil, errors.New("not used")
}

func (validationAI) TailorResume(context.Context, string, model.ResumeDocument, string, string, string, string) (*model.ResumeDocument, error) {
	return nil, errors.New("not used")
}

func TestAIHandlerValidationAndSettings(t *testing.T) {
	db := setupTestDB(t)
	cfgRepo := repository.NewConfigRepo(db)
	if err := cfgRepo.Create(context.Background(), &model.Config{ID: 1}); err != nil {
		t.Fatalf("create config: %v", err)
	}

	t.Run("validate key checks request and returns provider result", func(t *testing.T) {
		handler := NewAIHandler(cfgRepo, validationAI{valid: true, models: 3})
		router := newTestHandlerRouter(http.MethodPost, "/validate", handler.ValidateKey)

		for _, body := range []string{"", `{}`, `{"provider":"openrouter"}`} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("body %q status = %d, want %d", body, w.Code, http.StatusBadRequest)
			}
		}

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"provider":"openrouter","api_key":"key"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("validate status = %d: %s", w.Code, w.Body.String())
		}
		var response model.AIValidateKeyResponse
		parseJSON(t, w.Body.String(), &response)
		if !response.Valid || response.Models != 3 {
			t.Fatalf("validate response = %+v", response)
		}
	})

	t.Run("validation provider failures are reported without an HTTP failure", func(t *testing.T) {
		handler := NewAIHandler(cfgRepo, validationAI{err: errors.New("provider unavailable")})
		router := newTestHandlerRouter(http.MethodPost, "/validate", handler.ValidateKey)
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/validate", strings.NewReader(`{"provider":"openrouter","api_key":"key"}`))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		var response model.AIValidateKeyResponse
		parseJSON(t, w.Body.String(), &response)
		if w.Code != http.StatusOK || response.Valid || response.Error != "provider unavailable" {
			t.Fatalf("validation failure = %d, %+v", w.Code, response)
		}
	})

	t.Run("settings expose key presence but never the key and support partial updates", func(t *testing.T) {
		handler := NewAIHandler(cfgRepo, validationAI{})
		router := newTestHandlerRouter(http.MethodGet, "/settings", handler.GetSettings)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/settings", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("get default settings = %d: %s", w.Code, w.Body.String())
		}

		putRouter := newTestHandlerRouter(http.MethodPut, "/settings", handler.UpdateSettings)
		w = httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPut, "/settings", strings.NewReader(`{"provider":"openrouter","api_key":"secret","enabled":true,"user_name":"Ada","user_email":"ada@example.test"}`))
		req.Header.Set("Content-Type", "application/json")
		putRouter.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("update settings = %d: %s", w.Code, w.Body.String())
		}

		w = httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/settings", nil))
		var response model.AIInfoResponse
		parseJSON(t, w.Body.String(), &response)
		if !response.HasAPIKey || !response.Enabled || response.UserName != "Ada" || response.UserEmail != "ada@example.test" || strings.Contains(w.Body.String(), "secret") {
			t.Fatalf("unexpected settings response: %s", w.Body.String())
		}

		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodPut, "/settings", strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		putRouter.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("empty update = %d: %s", w.Code, w.Body.String())
		}
	})
}
