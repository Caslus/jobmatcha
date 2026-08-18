package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/caslus/jobmatcha/internal/model"
)

func TestSettingsAPI(t *testing.T) {
	db := setupTestDB(t)
	router, _ := setupTestRouter(db)

	// Get a session token
	token := authenticate(t, router, db)

	// 1. GET without auth
	t.Run("GetNoAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/settings", nil)
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 2. GET with auth — default empty settings
	t.Run("GetDefault", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/settings", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.SettingsResponse
		parseJSON(t, w.Body.String(), &resp)
		if len(resp.IncludeKeywords) != 0 {
			t.Errorf("expected empty include_keywords, got %v", resp.IncludeKeywords)
		}
		if len(resp.ExcludeKeywords) != 0 {
			t.Errorf("expected empty exclude_keywords, got %v", resp.ExcludeKeywords)
		}
		if len(resp.LocationKeywords) != 0 {
			t.Errorf("expected empty location_keywords, got %v", resp.LocationKeywords)
		}
		if resp.EmploymentType != "" {
			t.Errorf("expected empty employment_type, got '%s'", resp.EmploymentType)
		}
	})

	// 3. PUT without auth
	t.Run("PutNoAuth", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"include_keywords":["engineer","go"]}`
		req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	// 4. PUT include_keywords
	t.Run("PutInclude", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"include_keywords":["engineer","go","backend"]}`
		req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify via GET
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/settings", nil)
		req2.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w2, req2)

		if w2.Code != 200 {
			t.Fatalf("expected 200 on GET, got %d", w2.Code)
		}

		var resp model.SettingsResponse
		parseJSON(t, w2.Body.String(), &resp)
		if len(resp.IncludeKeywords) != 3 {
			t.Fatalf("expected 3 include_keywords, got %v", resp.IncludeKeywords)
		}
		if resp.IncludeKeywords[0] != "engineer" {
			t.Errorf("expected 'engineer', got '%s'", resp.IncludeKeywords[0])
		}
		if resp.IncludeKeywords[1] != "go" {
			t.Errorf("expected 'go', got '%s'", resp.IncludeKeywords[1])
		}
		if resp.IncludeKeywords[2] != "backend" {
			t.Errorf("expected 'backend', got '%s'", resp.IncludeKeywords[2])
		}
	})

	// 5. PUT exclude_keywords (partial update — should not affect include)
	t.Run("PutExclude", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"exclude_keywords":["frontend","manager"]}`
		req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		// Verify include preserved + exclude added
		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/settings", nil)
		req2.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w2, req2)

		var resp model.SettingsResponse
		parseJSON(t, w2.Body.String(), &resp)
		if len(resp.IncludeKeywords) != 3 {
			t.Errorf("expected 3 include_keywords preserved, got %v", resp.IncludeKeywords)
		}
		if len(resp.ExcludeKeywords) != 2 {
			t.Fatalf("expected 2 exclude_keywords, got %v", resp.ExcludeKeywords)
		}
		if resp.ExcludeKeywords[0] != "frontend" {
			t.Errorf("expected 'frontend', got '%s'", resp.ExcludeKeywords[0])
		}
		if resp.ExcludeKeywords[1] != "manager" {
			t.Errorf("expected 'manager', got '%s'", resp.ExcludeKeywords[1])
		}
	})

	// 6. PUT location_keywords
	t.Run("PutLocation", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"location_keywords":["tokyo","remote"]}`
		req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/settings", nil)
		req2.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w2, req2)

		var resp model.SettingsResponse
		parseJSON(t, w2.Body.String(), &resp)
		if len(resp.LocationKeywords) != 2 {
			t.Fatalf("expected 2 location_keywords, got %v", resp.LocationKeywords)
		}
		if resp.LocationKeywords[0] != "tokyo" {
			t.Errorf("expected 'tokyo', got '%s'", resp.LocationKeywords[0])
		}
		if resp.LocationKeywords[1] != "remote" {
			t.Errorf("expected 'remote', got '%s'", resp.LocationKeywords[1])
		}
	})

	// 7. PUT with empty body — should 400
	t.Run("PutEmptyBody", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{}`
		req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 400 {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}

		var resp model.ErrorResponse
		parseJSON(t, w.Body.String(), &resp)
		if resp.Error != "no fields to update" {
			t.Errorf("expected 'no fields to update', got '%s'", resp.Error)
		}
	})

	// 8. PUT overwrite include_keywords (sending nil should not clear if not included)
	t.Run("PutOverwriteInclude", func(t *testing.T) {
		w := httptest.NewRecorder()
		body := `{"include_keywords":["rust"]}`
		req := httptest.NewRequest("PUT", "/api/settings", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w, req)

		if w.Code != 200 {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}

		w2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("GET", "/api/settings", nil)
		req2.AddCookie(&http.Cookie{Name: "session", Value: token})
		router.ServeHTTP(w2, req2)

		var resp model.SettingsResponse
		parseJSON(t, w2.Body.String(), &resp)
		if len(resp.IncludeKeywords) != 1 || resp.IncludeKeywords[0] != "rust" {
			t.Errorf("expected ['rust'], got %v", resp.IncludeKeywords)
		}
		// Exclude and location should still be there from previous tests
		if len(resp.ExcludeKeywords) != 2 {
			t.Errorf("expected exclude_keywords preserved, got %v", resp.ExcludeKeywords)
		}
		if len(resp.LocationKeywords) != 2 {
			t.Errorf("expected location_keywords preserved, got %v", resp.LocationKeywords)
		}
	})
}