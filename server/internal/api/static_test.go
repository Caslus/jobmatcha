package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterStaticRoutes(t *testing.T) {
	dir := t.TempDir()
	assetsDir := filepath.Join(dir, "assets")
	if err := os.Mkdir(assetsDir, 0755); err != nil {
		t.Fatalf("creating assets directory: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>app</html>"), 0644); err != nil {
		t.Fatalf("writing index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(assetsDir, "app.js"), []byte("console.log('app')"), 0644); err != nil {
		t.Fatalf("writing asset: %v", err)
	}

	router := newTestRouter(t)
	if err := RegisterStaticRoutes(router, dir); err != nil {
		t.Fatalf("registering static routes: %v", err)
	}

	t.Run("serves asset", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))
		if w.Code != http.StatusOK || w.Body.String() != "console.log('app')" {
			t.Fatalf("asset response = %d %q", w.Code, w.Body.String())
		}
	})

	t.Run("falls back to the SPA", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/dashboard", nil))
		if w.Code != http.StatusOK || w.Body.String() != "<html>app</html>" {
			t.Fatalf("SPA response = %d %q", w.Code, w.Body.String())
		}
	})

	t.Run("does not mask API misses", func(t *testing.T) {
		w := httptest.NewRecorder()
		router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/missing", nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("API miss status = %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func newTestRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return gin.New()
}
