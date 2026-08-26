package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/caslus/jobmatcha/migrations"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "app.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get test sql db: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	if err := migrations.Migrate(db); err != nil {
		t.Fatalf("migrate test db: %v", err)
	}
	return db
}

func setupTestRouter(t *testing.T, db *gorm.DB, secure bool) (*gin.Engine, string) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	t.Setenv("COOKIE_SECURE", strconv.FormatBool(secure))
	secretFile := filepath.Join(t.TempDir(), "initial-password")
	repos := repository.NewRepositories(db)
	authSvc := service.NewAuthService(repos, secretFile)
	if err := authSvc.EnsureBootstrap(context.Background()); err != nil {
		t.Fatalf("bootstrap auth: %v", err)
	}
	r := gin.New()
	if secure {
		r.GET("/cookie", func(c *gin.Context) { setSessionCookie(c, "test", true) })
	}
	RegisterRoutes(r, repos, db, authSvc)
	return r, secretFile
}

func parseJSON(t *testing.T, body string, v interface{}) {
	t.Helper()
	if err := json.Unmarshal([]byte(body), v); err != nil {
		t.Fatalf("parse JSON: %v\nbody: %s", err, body)
	}
}

func bootstrapPassword(t *testing.T, file string) string {
	t.Helper()
	contents, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("read bootstrap password: %v", err)
	}
	info, err := os.Stat(file)
	if err != nil {
		t.Fatalf("stat bootstrap password: %v", err)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0600 {
		t.Fatalf("bootstrap file permissions = %o, want 600", info.Mode().Perm())
	}
	return strings.TrimSpace(string(contents))
}

func login(t *testing.T, router *gin.Engine, password string) *http.Cookie {
	t.Helper()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(`{"password":"`+password+`"}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login = %d: %s", w.Code, w.Body.String())
	}
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == sessionCookie {
			return cookie
		}
	}
	t.Fatal("login did not set session cookie")
	return nil
}

func authenticate(t *testing.T, router *gin.Engine, db *gorm.DB) string {
	t.Helper()
	repos := repository.NewRepositories(db)
	hash, err := bcrypt.GenerateFromPassword([]byte("test-password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash test password: %v", err)
	}
	if err := repos.Config.UpdateMap(context.Background(), map[string]interface{}{"password_hash": string(hash)}); err != nil {
		t.Fatalf("set test password: %v", err)
	}
	return login(t, router, "test-password").Value
}

func TestAuthBootstrapAndCookieSession(t *testing.T) {
	db := setupTestDB(t)
	router, secretFile := setupTestRouter(t, db, true)
	password := bootstrapPassword(t, secretFile)
	if password == "" {
		t.Fatal("bootstrap password is empty")
	}

	t.Run("status is public and unauthenticated", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/auth/status", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("status = %d", w.Code)
		}
		var status model.AuthStatusResponse
		parseJSON(t, w.Body.String(), &status)
		if status.Authenticated || status.SetupComplete {
			t.Fatalf("unexpected public status: %+v", status)
		}
	})

	t.Run("bearer tokens are rejected", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
		req.Header.Set("Authorization", "Bearer ignored")
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("bearer request = %d", w.Code)
		}
	})

	cookie := login(t, router, password)
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("insecure session cookie: %+v", cookie)
	}

	t.Run("cookie authenticates protected endpoints", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("protected endpoint = %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("password change removes bootstrap secret", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(`{"current_password":"`+password+`","new_password":"changed-password"}`))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("change password = %d: %s", w.Code, w.Body.String())
		}
		if _, err := os.Stat(secretFile); !os.IsNotExist(err) {
			t.Fatalf("bootstrap password file still exists: %v", err)
		}
	})

	t.Run("logout invalidates session", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("logout = %d", w.Code)
		}
		w = httptest.NewRecorder()
		req = httptest.NewRequest(http.MethodGet, "/api/settings", nil)
		req.AddCookie(cookie)
		router.ServeHTTP(w, req)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("logged-out endpoint = %d", w.Code)
		}
	})
}

func TestCookieSecureDefaults(t *testing.T) {
	if !CookieSecureFromEnv("") || CookieSecureFromEnv("false") {
		t.Fatal("unexpected cookie secure configuration")
	}
}

func TestHealthEndpoint(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db, false)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("health = %d", w.Code)
	}
}
