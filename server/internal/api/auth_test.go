package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/migrations"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "jobmatcha-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp db: %v", err)
	}
	tmpFile.Close()
	t.Cleanup(func() { os.Remove(tmpFile.Name()) })

	db, err := gorm.Open(sqlite.Open(tmpFile.Name()), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := migrations.Migrate(db); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func setupTestRouter(db *gorm.DB) (*gin.Engine, *repository.Repositories) {
	gin.SetMode(gin.TestMode)
	repos := repository.NewRepositories(db)
	r := gin.New()
	RegisterRoutes(r, repos, db)
	return r, repos
}

func parseJSON(t *testing.T, body string, v interface{}) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("failed to parse JSON: %v\nbody: %s", err, body)
	}
}

func TestAuthFlow(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	// 1. Status before setup
	t.Run("StatusBeforeSetup", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/auth/status", nil)
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.AuthStatusResponse
		parseJSON(t, w.Body.String(), &resp)
		if resp.Authenticated {
			t.Error("expected not authenticated before setup")
		}
		if resp.SetupComplete {
			t.Error("expected setup not complete before setup")
		}
	})

	// 2. Setup with short password (should fail)
	t.Run("SetupShortPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"short"}`
		req := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Fatalf("expected 400 for short password, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 3. Setup with valid password
	t.Run("SetupValid", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"password123"}`
		req := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.AuthTokenResponse
		parseJSON(t, w.Body.String(), &resp)
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}
	})

	// 4. Setup again (should fail — already completed)
	t.Run("SetupAgain", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"password123"}`
		req := httptest.NewRequest("POST", "/api/auth/setup", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 409 {
			t.Fatalf("expected 409 for re-setup, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 5. Login with wrong password
	t.Run("LoginWrongPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"wrongpass"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401 for wrong password, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 6. Login with correct password
	var sessionToken string
	t.Run("LoginValid", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"password123"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.AuthTokenResponse
		parseJSON(t, w.Body.String(), &resp)
		if resp.Token == "" {
			t.Error("expected non-empty token")
		}
		sessionToken = resp.Token
	})

	// 7. Status with valid session cookie
	t.Run("StatusAuthenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/auth/status", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.AuthStatusResponse
		parseJSON(t, w.Body.String(), &resp)
		if !resp.Authenticated {
			t.Error("expected authenticated")
		}
		if !resp.SetupComplete {
			t.Error("expected setup complete")
		}
	})

	// 8. Protected route without session (should fail)
	t.Run("ProtectedNoAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"current_password":"password123","new_password":"newpass456"}`
		req := httptest.NewRequest("POST", "/api/auth/change-password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 9. Change password
	t.Run("ChangePassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"current_password":"password123","new_password":"newpass456"}`
		req := httptest.NewRequest("POST", "/api/auth/change-password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 10. Login with old password (should fail after change)
	t.Run("LoginOldPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"password123"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401 after password change, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 11. Login with new password
	t.Run("LoginNewPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"newpass456"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 12. Logout
	t.Run("Logout", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 13. Status after logout (should be unauthenticated)
	t.Run("StatusAfterLogout", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/auth/status", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.AuthStatusResponse
		parseJSON(t, w.Body.String(), &resp)
		if resp.Authenticated {
			t.Error("expected not authenticated after logout")
		}
	})
}

func TestHealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	parseJSON(t, w.Body.String(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %v", resp)
	}
}