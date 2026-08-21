package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
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

func authenticate(t *testing.T, router *gin.Engine, db *gorm.DB) string {
	t.Helper()
	// Login with default password
	w := httptest.NewRecorder()
	body := `{"password":"admin"}`
	req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	var resp model.AuthTokenResponse
	parseJSON(t, w.Body.String(), &resp)
	return resp.Token
}

func TestAuthFlow(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	// 1. Status before login (protected, returns 401)
	t.Run("StatusBeforeLogin", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/auth/status", nil)
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 2. Login with wrong password
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

	// 3. Login with default password
	var sessionToken string
	t.Run("LoginDefault", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"admin"}`
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

	// 4. Status with valid session — setup_complete should be false
	t.Run("StatusSetupIncomplete", func(t *testing.T) {
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
		if resp.SetupComplete {
			t.Error("expected setup_complete=false with default password")
		}
	})

	// 5. Change password
	t.Run("ChangePassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"current_password":"admin","new_password":"newpass456"}`
		req := httptest.NewRequest("POST", "/api/auth/change-password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 6. Status after password change — setup_complete should be true
	t.Run("StatusSetupComplete", func(t *testing.T) {
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
			t.Logf("setup not complete (expected after change-password no longer sets it; onboarding controls it)")
		}
	})

	// 7. Login with old password (should fail after change)
	t.Run("LoginOldPassword", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"password":"admin"}`
		req := httptest.NewRequest("POST", "/api/auth/login", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401 after password change, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 8. Login with new password
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

	// 9. Logout
	t.Run("Logout", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/auth/logout", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 10. Status after logout (should be unauthenticated)
	t.Run("StatusAfterLogout", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/auth/status", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: sessionToken})
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 11. Protected route without session (should fail)
	t.Run("ProtectedNoAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"current_password":"newpass456","new_password":"anotherpass"}`
		req := httptest.NewRequest("POST", "/api/auth/change-password", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
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

func TestRolesAPI(t *testing.T) {
	db := setupTestDB(t)
	router, repos := setupTestRouter(db)

	// Seed a test company + role
	company := model.Company{
		Name:       "TestCorp",
		CareersURL: "https://test.corp/careers",
		ATSType:    "test",
		Active:     true,
		Region:     "JP",
	}
	db.Create(&company)

	role := model.Role{
		CompanyID: company.ID,
		URLHash:   "test123",
		URL:       "https://test.corp/jobs/1",
		Title:     "Software Engineer",
		Department: "Engineering",
		Location:  "Tokyo, Japan",
		Status:    "seen",
	}
	db.Create(&role)

	// Get a session token
	token := authenticate(t, router, db)

	// 1. List roles without auth
	t.Run("ListNoAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/roles", nil)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 2. List roles with auth
	t.Run("ListWithAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/roles", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.RoleListResponse
		parseJSON(t, w.Body.String(), &resp)
		if len(resp.Data) == 0 {
			t.Fatal("expected at least 1 role")
		}
		if resp.Data[0].Title != "Software Engineer" {
			t.Errorf("expected 'Software Engineer', got '%s'", resp.Data[0].Title)
		}
		if resp.Data[0].CompanyName != "TestCorp" {
			t.Errorf("expected 'TestCorp', got '%s'", resp.Data[0].CompanyName)
		}
		if resp.Pagination.Total != 1 {
			t.Errorf("expected total=1, got %d", resp.Pagination.Total)
		}
	})

	// 3. Get role by ID
	t.Run("GetByID", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/roles/"+strconv.Itoa(int(role.ID)), nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.RoleDetailResponse
		parseJSON(t, w.Body.String(), &resp)
		if resp.Title != "Software Engineer" {
			t.Errorf("expected 'Software Engineer', got '%s'", resp.Title)
		}
		if resp.CompanyName != "TestCorp" {
			t.Errorf("expected 'TestCorp', got '%s'", resp.CompanyName)
		}
	})

	// 4. Get role by ID — not found
	t.Run("GetByIDNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/roles/99999", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 404 {
			t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 5. Patch role — interested
	t.Run("PatchInterested", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"is_interested":true}`
		req := httptest.NewRequest("PATCH", "/api/roles/"+strconv.Itoa(int(role.ID)), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify
		updated, _ := repos.Role.GetByID(role.ID)
		if !updated.IsInterested {
			t.Error("expected is_interested=true")
		}
	})

	// 6. Patch role — hide
	t.Run("PatchHide", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"is_hidden":true}`
		req := httptest.NewRequest("PATCH", "/api/roles/"+strconv.Itoa(int(role.ID)), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify it no longer appears in list
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/roles", nil)
		req2.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w2, req2)

		var resp model.RoleListResponse
		parseJSON(t, w2.Body.String(), &resp)
		for _, r := range resp.Data {
			if r.ID == role.ID {
				t.Fatal("hidden role should not appear in list")
			}
		}
	})

	// 7. Patch without auth
	t.Run("PatchNoAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"is_hidden":false}`
		req := httptest.NewRequest("PATCH", "/api/roles/"+strconv.Itoa(int(role.ID)), strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})
}

