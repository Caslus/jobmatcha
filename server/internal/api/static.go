package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterStaticRoutes serves the built SPA while keeping all API paths under
// /api reserved for the JSON API. staticDir must contain the SPA's index.html.
func RegisterStaticRoutes(r *gin.Engine, staticDir string) error {
	indexPath := filepath.Join(staticDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		return fmt.Errorf("checking frontend index: %w", err)
	}

	r.StaticFS("/assets", http.Dir(filepath.Join(staticDir, "assets")))
	r.StaticFile("/favicon.svg", filepath.Join(staticDir, "favicon.svg"))
	r.NoRoute(func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found."})
			return
		}
		if c.Request.Method != http.MethodGet && c.Request.Method != http.MethodHead {
			c.Status(http.StatusNotFound)
			return
		}

		c.File(indexPath)
	})

	return nil
}
