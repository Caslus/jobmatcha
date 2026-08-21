package api

import (
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, repos *repository.Repositories, db *gorm.DB) *service.SchedulerService {
	scannerSvc := service.NewScannerService(db, repos)
	schedulerSvc := service.NewSchedulerService(repos.Config, scannerSvc)
	schedulerSvc.Start()

	auth := NewAuthHandler(repos.Config, db)
	roles := NewRoleHandler(repos)
	settings := NewSettingsHandler(repos.Config, schedulerSvc)
	scan := NewScanHandler(scannerSvc)
	aiH := NewAIHandler(repos.Config)
	onboarding := NewOnboardingHandler(repos.Config, scannerSvc, schedulerSvc)

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

		// AI & Onboarding endpoints
		protected.POST("/ai/validate-key", aiH.ValidateKey)
		protected.POST("/ai/parse-resume", aiH.ParseResume)
		protected.GET("/settings/ai", aiH.GetSettings)
		protected.PUT("/settings/ai", aiH.UpdateSettings)
		protected.POST("/onboarding/complete", onboarding.Complete)
	}

	return schedulerSvc
}