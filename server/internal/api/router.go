package api

import (
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, repos *repository.Repositories, db *gorm.DB) {
	auth := NewAuthHandler(repos.Config, db)
	roles := NewRoleHandler(repos)
	settings := NewSettingsHandler(repos.Config)
	scan := NewScanHandler(service.NewScannerService(db, repos))

	// Public routes (only what's needed before auth)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/api/auth/login", auth.Login)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(Authenticated(db))
	{
		protected.POST("/auth/logout", auth.Logout)
		protected.GET("/auth/status", auth.Status)
		protected.POST("/auth/change-password", auth.ChangePassword)
		protected.GET("/roles", roles.List)
		protected.GET("/roles/:id", roles.GetByID)
		protected.PATCH("/roles/:id", roles.Patch)
		protected.GET("/settings", settings.Get)
		protected.PUT("/settings", settings.Update)
		protected.POST("/scan", scan.Start)
		protected.GET("/scan/latest", scan.GetLatest)
		protected.GET("/scan/:id", scan.GetByID)
	}
}