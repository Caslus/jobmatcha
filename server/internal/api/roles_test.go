package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
)

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
}
