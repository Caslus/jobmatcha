package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

type companyBoardAdapterFake struct{}

func (companyBoardAdapterFake) SupportsAdapter(adapter string) bool { return adapter == "fake" }

func (companyBoardAdapterFake) NormalizeCareerBoard(_ context.Context, identity model.BoardIdentity) (model.BoardIdentity, error) {
	if identity.Provider != "fake" || identity.BoardIdentifier == "invalid" {
		return model.BoardIdentity{}, fmt.Errorf("invalid career board")
	}
	return model.BoardIdentity{
		Provider:        identity.Provider,
		BoardIdentifier: identity.BoardIdentifier,
		CanonicalURL:    "https://boards.example/" + identity.BoardIdentifier,
	}, nil
}

type careerBoardDiscovererFake struct {
	result *model.CareerBoardDiscoveryResponse
	err    error
}

func (f careerBoardDiscovererFake) DiscoverCareerBoards(_ context.Context, _ string) (*model.CareerBoardDiscoveryResponse, error) {
	return f.result, f.err
}

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
	boards := []model.CareerBoard{
		{CompanyID: companies[0].ID, Provider: "workable", BoardIdentifier: "alpha-one", CanonicalURL: "https://apply.workable.com/alpha-one/", Active: true},
		{CompanyID: companies[0].ID, Provider: "workable", BoardIdentifier: "alpha-two", CanonicalURL: "https://apply.workable.com/alpha-two/", Active: true},
	}
	if err := db.Create(&boards).Error; err != nil {
		t.Fatalf("create boards: %v", err)
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
	if len(list.Data) != 2 || list.Data[0].Name != "Alpha" || list.Data[0].RoleCount != 2 || list.Data[0].BoardCount != 2 || len(list.Data[0].CareerBoards) != 2 {
		t.Fatalf("list = %#v", list)
	}
	w = request(http.MethodPatch, "/api/companies/1", `{"active":false}`)
	if w.Code != http.StatusOK {
		t.Fatalf("single update = %d: %s", w.Code, w.Body.String())
	}
	w = request(http.MethodPatch, "/api/companies/1/boards/"+strconv.FormatUint(uint64(boards[0].ID), 10), `{"active":false}`)
	if w.Code != http.StatusOK {
		t.Fatalf("board update = %d: %s", w.Code, w.Body.String())
	}
	var board model.CareerBoard
	if err := db.First(&board, boards[1].ID).Error; err != nil || !board.Active {
		t.Fatalf("sibling board changed: %#v, %v", board, err)
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
	var before int64
	if err := db.Model(&model.Company{}).Count(&before).Error; err != nil {
		t.Fatal(err)
	}
	w = request(http.MethodPost, "/api/companies/register", `{"candidates":[]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("zero registration = %d: %s", w.Code, w.Body.String())
	}
	var after int64
	if err := db.Model(&model.Company{}).Count(&after).Error; err != nil || after != before {
		t.Fatalf("zero selection wrote companies: %d -> %d, %v", before, after, err)
	}
	w = request(http.MethodPost, "/api/companies/register", `{"candidates":[{"company_name":"Selected","careers_url":"https://selected.test/careers","provider":"workable","board_identifier":"selected","canonical_url":"https://apply.workable.com/selected/"}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("registration = %d: %s", w.Code, w.Body.String())
	}
	var selected model.Company
	if err := db.Where("name = ?", "Selected").First(&selected).Error; err != nil {
		t.Fatalf("selected company: %v", err)
	}
	var selectedBoards []model.CareerBoard
	if err := db.Where("company_id = ?", selected.ID).Find(&selectedBoards).Error; err != nil || len(selectedBoards) != 1 || !selectedBoards[0].Active {
		t.Fatalf("selected boards = %#v, %v", selectedBoards, err)
	}
	if err := db.Model(&model.Company{}).Count(&before).Error; err != nil {
		t.Fatal(err)
	}
	w = request(http.MethodPost, "/api/companies/register", `{"candidates":[{"company_name":"Duplicate employer","careers_url":"https://duplicate.test/careers","provider":"workable","board_identifier":"selected","canonical_url":"https://apply.workable.com/selected/"}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("duplicate registration = %d: %s", w.Code, w.Body.String())
	}
	if err := db.Model(&model.Company{}).Count(&after).Error; err != nil || after != before {
		t.Fatalf("duplicate selection wrote company: %d -> %d, %v", before, after, err)
	}
	w = request(http.MethodPost, "/api/companies/register", `{"candidates":[{"company_name":"Different group company","careers_url":"https://group.test/careers","provider":"workable","board_identifier":"alpha-one","canonical_url":"https://apply.workable.com/alpha-one/"},{"company_name":"Another group company","careers_url":"https://group.test/careers","provider":"workable","board_identifier":"alpha-three","canonical_url":"https://apply.workable.com/alpha-three/"}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("group registration = %d: %s", w.Code, w.Body.String())
	}
	if err := db.Model(&model.Company{}).Count(&after).Error; err != nil || after != before {
		t.Fatalf("group registration wrote company: %d -> %d, %v", before, after, err)
	}
	var groupBoard model.CareerBoard
	if err := db.Where("provider = ? AND board_identifier = ?", "workable", "alpha-three").First(&groupBoard).Error; err != nil || groupBoard.CompanyID != companies[0].ID {
		t.Fatalf("group board = %#v, %v", groupBoard, err)
	}
	w = request(http.MethodPost, "/api/companies/register", `{"candidates":[{"company_name":"Separate company","careers_url":"https://group.test/careers","provider":"workable","board_identifier":"separate-board","canonical_url":"https://apply.workable.com/separate-board/","separate":true}]}`)
	if w.Code != http.StatusOK {
		t.Fatalf("separate registration = %d: %s", w.Code, w.Body.String())
	}
	var separateCompany model.Company
	if err := db.Where("name = ?", "Separate company").First(&separateCompany).Error; err != nil {
		t.Fatalf("separate company: %v", err)
	}
	var separateBoard model.CareerBoard
	if err := db.Where("provider = ? AND board_identifier = ?", "workable", "separate-board").First(&separateBoard).Error; err != nil || separateBoard.CompanyID != separateCompany.ID {
		t.Fatalf("separate board = %#v, %v", separateBoard, err)
	}
}

func TestCompanyManagementHandlers(t *testing.T) {
	db := setupTestDB(t)
	company := model.Company{Name: "Original", CareersURL: "https://original.example/careers", Active: true}
	other := model.Company{Name: "Other", CareersURL: "https://other.example/careers", Active: true}
	if err := db.Create(&[]model.Company{company, other}).Error; err != nil {
		t.Fatalf("create companies: %v", err)
	}
	if err := db.Where("name = ?", "Original").First(&company).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Where("name = ?", "Other").First(&other).Error; err != nil {
		t.Fatal(err)
	}

	companyService := service.NewCompanyService(repository.NewRepositories(db).Company, companyBoardAdapterFake{})
	handler := NewCompanyHandler(companyService, careerBoardDiscovererFake{result: &model.CareerBoardDiscoveryResponse{Candidates: []model.CareerBoardDiscoveryCandidate{{Provider: "fake", BoardIdentifier: "primary", CanonicalURL: "https://boards.example/primary"}}}})
	router := gin.New()
	router.POST("/companies/discover", handler.Discover)
	router.PATCH("/companies/:id/details", handler.UpdateDetails)
	router.DELETE("/companies/:id", handler.Delete)
	router.POST("/companies/:id/boards", handler.CreateBoard)
	router.PATCH("/companies/:id/boards/:boardID/details", handler.UpdateBoardDetails)
	router.DELETE("/companies/:id/boards/:boardID", handler.DeleteBoard)
	request := func(method, path, body string) *httptest.ResponseRecorder {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(method, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		return w
	}

	if w := request(http.MethodPost, "/companies/discover", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("missing careers URL = %d", w.Code)
	}
	if w := request(http.MethodPost, "/companies/discover", `{"careers_url":"https://original.example/careers"}`); w.Code != http.StatusOK {
		t.Fatalf("discover = %d: %s", w.Code, w.Body.String())
	}
	if w := request(http.MethodPatch, "/companies/invalid/details", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid company ID = %d", w.Code)
	}
	if w := request(http.MethodPatch, "/companies/1/details", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("missing company name = %d", w.Code)
	}
	if w := request(http.MethodPatch, "/companies/1/details", `{"name":" Renamed ","location":" Tokyo "}`); w.Code != http.StatusOK {
		t.Fatalf("update company = %d: %s", w.Code, w.Body.String())
	}
	if w := request(http.MethodPost, "/companies/1/boards", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("missing board fields = %d", w.Code)
	}
	w := request(http.MethodPost, "/companies/1/boards", `{"provider":"fake","board_identifier":"primary","canonical_url":"https://ignored.example"}`)
	if w.Code != http.StatusCreated {
		t.Fatalf("create board = %d: %s", w.Code, w.Body.String())
	}
	var item model.CompanyListItem
	parseJSON(t, w.Body.String(), &item)
	if item.Name != "Renamed" || item.Location != "Tokyo" || len(item.CareerBoards) != 1 {
		t.Fatalf("created board company = %#v", item)
	}
	boardID := item.CareerBoards[0].ID
	if w := request(http.MethodPatch, "/companies/1/boards/invalid/details", `{}`); w.Code != http.StatusBadRequest {
		t.Fatalf("invalid board ID = %d", w.Code)
	}
	if w := request(http.MethodPatch, "/companies/2/boards/"+strconv.FormatUint(uint64(boardID), 10)+"/details", `{"provider":"fake","board_identifier":"wrong-owner","canonical_url":"https://ignored.example"}`); w.Code != http.StatusNotFound {
		t.Fatalf("wrong board owner = %d: %s", w.Code, w.Body.String())
	}
	w = request(http.MethodPatch, "/companies/1/boards/"+strconv.FormatUint(uint64(boardID), 10)+"/details", `{"provider":"fake","board_identifier":"replacement","canonical_url":"https://ignored.example"}`)
	if w.Code != http.StatusOK {
		t.Fatalf("update board = %d: %s", w.Code, w.Body.String())
	}
	if w := request(http.MethodDelete, "/companies/2/boards/"+strconv.FormatUint(uint64(boardID), 10), ""); w.Code != http.StatusNotFound {
		t.Fatalf("delete wrong board owner = %d", w.Code)
	}
	if w := request(http.MethodDelete, "/companies/1/boards/"+strconv.FormatUint(uint64(boardID), 10), ""); w.Code != http.StatusOK {
		t.Fatalf("delete board = %d: %s", w.Code, w.Body.String())
	}
	if err := db.Create(&model.Role{CompanyID: company.ID, URLHash: "historical-role", URL: "https://roles.example/1", Title: "Historical role"}).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}
	if w := request(http.MethodDelete, "/companies/1", ""); w.Code != http.StatusNoContent {
		t.Fatalf("delete company = %d: %s", w.Code, w.Body.String())
	}
	var role model.Role
	if err := db.Where("url_hash = ?", "historical-role").First(&role).Error; err != nil {
		t.Fatalf("company deletion removed historical role: %v", err)
	}
	if w := request(http.MethodDelete, "/companies/1", ""); w.Code != http.StatusNotFound {
		t.Fatalf("delete missing company = %d", w.Code)
	}

	unavailableRouter := gin.New()
	unavailableRouter.POST("/companies/discover", NewCompanyHandler(companyService, nil).Discover)
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/companies/discover", strings.NewReader(`{"careers_url":"https://original.example/careers"}`))
	req.Header.Set("Content-Type", "application/json")
	unavailableRouter.ServeHTTP(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("unavailable discovery = %d", w.Code)
	}

	failingRouter := gin.New()
	failingRouter.POST("/companies/discover", NewCompanyHandler(companyService, careerBoardDiscovererFake{err: fmt.Errorf("fetch failed")}).Discover)
	w = httptest.NewRecorder()
	failingRouter.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("failed discovery = %d", w.Code)
	}
}
