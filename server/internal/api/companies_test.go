package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
)

func TestCompaniesAPIListsAndUpdatesAuthenticatedCompanies(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(t, db, false)
	companies := []model.Company{{Name: "Alpha", CareersURL: "https://alpha.test", ATSType: "workable", Active: true}, {Name: "Beta", CareersURL: "https://beta.test", ATSType: "unsupported", Active: false}}
	if err := db.Create(&companies).Error; err != nil {
		t.Fatalf("create companies: %v", err)
	}
	if err := db.Create(&[]model.Role{
		{CompanyID: companies[0].ID, URLHash: "alpha-1", URL: "https://alpha.test/1", Title: "Alpha role"},
		{CompanyID: companies[0].ID, URLHash: "alpha-2", URL: "https://alpha.test/2", Title: "Another alpha role"},
	}).Error; err != nil {
		t.Fatalf("create roles: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/companies", nil))
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous list = %d", w.Code)
	}
	token := authenticate(t, router, db)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: sessionCookie, Value: token})
		router.ServeHTTP(w, req)
		return w
	}
	w = request(http.MethodGet, "/api/companies", "")
	if w.Code != http.StatusOK {
		t.Fatalf("list = %d: %s", w.Code, w.Body.String())
	}
	var list model.CompanyListResponse
	parseJSON(t, w.Body.String(), &list)
	if len(list.Data) != 2 || list.Data[0].Name != "Alpha" || list.Data[0].RoleCount != 2 || list.Data[1].AdapterStatus != "unsupported" {
		t.Fatalf("list = %#v", list)
	}
	w = request(http.MethodPatch, "/api/companies/1", `{"active":false}`)
	if w.Code != http.StatusOK {
		t.Fatalf("single update = %d: %s", w.Code, w.Body.String())
	}
	w = request(http.MethodPut, "/api/companies/active", `{"company_ids":[1,2],"active":true}`)
	if w.Code != http.StatusOK {
		t.Fatalf("bulk update = %d: %s", w.Code, w.Body.String())
	}
	for _, body := range []string{`{}`, `{"company_ids":[],"active":true}`, `{"company_ids":[1,1],"active":true}`} {
		if w = request(http.MethodPut, "/api/companies/active", body); w.Code != http.StatusBadRequest {
			t.Fatalf("invalid bulk %s = %d", body, w.Code)
		}
	}
	w = request(http.MethodPut, "/api/companies/active", `{"company_ids":[1,999],"active":false}`)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing bulk = %d", w.Code)
	}
	var alpha model.Company
	if err := db.First(&alpha, companies[0].ID).Error; err != nil || !alpha.Active {
		t.Fatalf("partial update: %#v, %v", alpha, err)
	}
}
