package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupRoleHandlerRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db := setupTestDB(t)
	if err := db.Create(&model.Config{ID: 1}).Error; err != nil {
		t.Fatalf("create config: %v", err)
	}

	handler := NewRoleHandler(repository.NewRepositories(db))
	router := gin.New()
	router.GET("/roles", handler.List)
	router.PATCH("/roles/:id", handler.Patch)
	return router, db
}

func TestRolesContractRequiresAuthenticationAndPersistsPreferences(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db, false)

	company := model.Company{Name: "Example", CareersURL: "https://example.test/careers"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	postedAt := time.Now().Add(-time.Hour)
	role := model.Role{CompanyID: company.ID, URLHash: "role-1", URL: "https://example.test/jobs/1", Title: "Go engineer", PostedAt: &postedAt}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/roles", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous roles status = %d, want %d", w.Code, http.StatusUnauthorized)
	}

	token := authenticate(t, router, db)
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/api/roles/"+"1", strings.NewReader(`{"is_interested":true}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("patch role status = %d: %s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/roles?page=1&per_page=1", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list roles status = %d: %s", w.Code, w.Body.String())
	}
	var response model.RoleListResponse
	parseJSON(t, w.Body.String(), &response)
	if len(response.Data) != 1 || !response.Data[0].IsInterested || response.Data[0].Title != role.Title {
		t.Fatalf("unexpected role response: %#v", response)
	}

	t.Run("detail handles invalid and missing IDs", func(t *testing.T) {
		for _, path := range []string{"/api/roles/not-a-number", "/api/roles/9999"} {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, path, nil)
			req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
			router.ServeHTTP(w, req)
			if path == "/api/roles/not-a-number" && w.Code != http.StatusBadRequest {
				t.Fatalf("invalid detail status = %d", w.Code)
			}
			if path == "/api/roles/9999" && w.Code != http.StatusNotFound {
				t.Fatalf("missing detail status = %d", w.Code)
			}
		}
	})

	t.Run("detail includes scored match analysis", func(t *testing.T) {
		if err := db.Model(&model.Config{}).Where("id = 1").Updates(map[string]interface{}{
			"include_keywords": model.StringSlice{"go", "engineer"},
		}).Error; err != nil {
			t.Fatalf("set relevance config: %v", err)
		}
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/roles/1", nil)
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("detail status = %d: %s", w.Code, w.Body.String())
		}
		var response model.RoleDetailResponse
		parseJSON(t, w.Body.String(), &response)
		if response.Title != role.Title || response.CompanyName != company.Name || response.MatchDetails == nil || response.MatchPercent <= 0 {
			t.Fatalf("unexpected role detail: %#v", response)
		}
	})
}

func TestRoleListFiltersSortsAndNormalizesPagination(t *testing.T) {
	router, db := setupRoleHandlerRouter(t)
	company := model.Company{Name: "Example", CareersURL: "https://example.test/careers"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}

	older := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	newer := older.Add(24 * time.Hour)
	roles := []model.Role{
		{CompanyID: company.ID, URLHash: "older", URL: "https://example.test/jobs/older", Title: "Older", PostedAt: &older},
		{CompanyID: company.ID, URLHash: "newer", URL: "https://example.test/jobs/newer", Title: "Newer", PostedAt: &newer},
		{CompanyID: company.ID, URLHash: "undated", URL: "https://example.test/jobs/undated", Title: "Undated"},
		{CompanyID: company.ID, URLHash: "hidden", URL: "https://example.test/jobs/hidden", Title: "Hidden", IsHidden: true},
	}
	if err := db.Create(&roles).Error; err != nil {
		t.Fatalf("create roles: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/roles?page=0&per_page=101", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("list roles = %d: %s", w.Code, w.Body.String())
	}
	var response model.RoleListResponse
	parseJSON(t, w.Body.String(), &response)
	if response.Pagination != (model.PaginationInfo{Total: 3, Page: 1, PerPage: 25, TotalPages: 1}) || response.TotalAll != 4 {
		t.Fatalf("unexpected pagination: %#v, total_all=%d", response.Pagination, response.TotalAll)
	}
	if len(response.Data) != 3 || response.Data[0].Title != "Newer" || response.Data[1].Title != "Older" || response.Data[2].Title != "Undated" {
		t.Fatalf("roles were not sorted and filtered as expected: %#v", response.Data)
	}

	if err := db.Model(&model.Config{}).Where("id = 1").Update("max_days_old", 1).Error; err != nil {
		t.Fatalf("set max age: %v", err)
	}
	recent := time.Now().Add(-time.Hour)
	if err := db.Create(&model.Role{CompanyID: company.ID, URLHash: "recent", URL: "https://example.test/jobs/recent", Title: "Recent", PostedAt: &recent}).Error; err != nil {
		t.Fatalf("create recent role: %v", err)
	}
	w = httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/roles?page=2&per_page=1", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("aged list roles = %d: %s", w.Code, w.Body.String())
	}
	parseJSON(t, w.Body.String(), &response)
	if len(response.Data) != 0 || response.Pagination != (model.PaginationInfo{Total: 1, Page: 2, PerPage: 1, TotalPages: 1}) || response.TotalAll != 5 {
		t.Fatalf("unexpected max-age response: %#v", response)
	}
}

func TestRoleListExcludesDisabledCompanyRolesAndRestoresThemWhenEnabled(t *testing.T) {
	router, db := setupRoleHandlerRouter(t)
	enabled := model.Company{Name: "Enabled", CareersURL: "https://enabled.test/careers", Active: true}
	disabled := model.Company{Name: "Disabled", CareersURL: "https://disabled.test/careers", Active: false}
	if err := db.Create(&enabled).Error; err != nil {
		t.Fatalf("create enabled company: %v", err)
	}
	if err := db.Create(&disabled).Error; err != nil {
		t.Fatalf("create disabled company: %v", err)
	}
	if err := db.Model(&disabled).Update("active", false).Error; err != nil {
		t.Fatalf("disable company: %v", err)
	}
	if err := db.Create(&[]model.Role{
		{CompanyID: enabled.ID, URLHash: "enabled-role", URL: "https://enabled.test/jobs/1", Title: "Visible"},
		{CompanyID: disabled.ID, URLHash: "disabled-role", URL: "https://disabled.test/jobs/1", Title: "Hidden by company"},
	}).Error; err != nil {
		t.Fatalf("create roles: %v", err)
	}

	list := func() model.RoleListResponse {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/roles", nil))
		if w.Code != http.StatusOK {
			t.Fatalf("list roles = %d: %s", w.Code, w.Body.String())
		}
		var response model.RoleListResponse
		parseJSON(t, w.Body.String(), &response)
		return response
	}

	response := list()
	if len(response.Data) != 1 || response.Data[0].Title != "Visible" || response.Pagination.Total != 1 || response.TotalAll != 1 {
		t.Fatalf("disabled company leaked into list: %#v", response)
	}
	if err := db.Model(&model.Company{}).Where("id = ?", disabled.ID).Update("active", true).Error; err != nil {
		t.Fatalf("enable company: %v", err)
	}
	response = list()
	if len(response.Data) != 2 || response.Pagination.Total != 2 || response.TotalAll != 2 {
		t.Fatalf("re-enabled company roles not restored: %#v", response)
	}
}

func TestRolePatchValidationAndUpdatesFalseValues(t *testing.T) {
	router, db := setupRoleHandlerRouter(t)
	company := model.Company{Name: "Example", CareersURL: "https://example.test/careers"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatalf("create company: %v", err)
	}
	role := model.Role{CompanyID: company.ID, URLHash: "role", URL: "https://example.test/jobs/role", Title: "Role", IsHidden: true, IsInterested: true}
	if err := db.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	for _, tc := range []struct {
		name string
		path string
		body string
	}{
		{name: "invalid ID", path: "/roles/not-a-number", body: `{}`},
		{name: "invalid JSON", path: "/roles/1", body: `{`},
		{name: "no fields", path: "/roles/1", body: `{}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPatch, tc.path, strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(w, req)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("patch = %d, want %d: %s", w.Code, http.StatusBadRequest, w.Body.String())
			}
		})
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPatch, "/roles/1", strings.NewReader(`{"is_hidden":false,"is_interested":false}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("patch = %d: %s", w.Code, w.Body.String())
	}
	var updated model.Role
	if err := db.First(&updated, role.ID).Error; err != nil {
		t.Fatalf("load updated role: %v", err)
	}
	if updated.IsHidden || updated.IsInterested {
		t.Fatalf("false values were not persisted: %#v", updated)
	}
}
