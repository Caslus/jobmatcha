package api

import (
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

type RouteDependencies struct {
	Auth         *service.AuthService
	Scanner      service.Scanner
	Scheduler    service.Scheduler
	AI           service.AIClient
	CookieSecure bool
}

// RegisterRoutes only binds routes. Background work is owned by the
// application constructor, rather than being started as a routing side effect.
func RegisterRoutes(r *gin.Engine, repos *repository.Repositories, deps RouteDependencies) {
	authSvc := deps.Auth
	scannerSvc := deps.Scanner
	schedulerSvc := deps.Scheduler

	auth := NewAuthHandler(authSvc, deps.CookieSecure)
	roles := NewRoleHandler(repos)
	settings := NewSettingsHandler(repos.Config, schedulerSvc)
	scan := NewScanHandler(scannerSvc)
	companies := NewCompanyHandler(service.NewCompanyService(repos.Company, scannerSvc))
	aiH := NewAIHandler(repos.Config, deps.AI)
	resumeH := NewResumeHandler(repos, deps.AI)
	onboarding := NewOnboardingHandler(repos.Config, scannerSvc, schedulerSvc)

	// Public routes (only what's needed before auth)
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/api/auth/login", auth.Login)
	r.GET("/api/auth/status", auth.Status)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(Authenticated(authSvc))
	{
		protected.POST("/auth/logout", auth.Logout)
		protected.POST("/auth/change-password", auth.ChangePassword)
		protected.GET("/roles", roles.List)
		protected.GET("/companies", companies.List)
		protected.PATCH("/companies/:id", companies.UpdateActive)
		protected.PUT("/companies/active", companies.UpdateActiveBulk)
		protected.GET("/roles/:id", roles.GetByID)
		protected.PATCH("/roles/:id", roles.Patch)
		protected.GET("/settings", settings.Get)
		protected.PUT("/settings", settings.Update)
		protected.POST("/scan", scan.Start)
		protected.GET("/scan/latest", scan.GetLatest)
		protected.GET("/scan/:id", scan.GetByID)

		// AI & Onboarding endpoints
		protected.POST("/ai/validate-key", aiH.ValidateKey)
		protected.POST("/ai/parse-resume", resumeH.ParseUpload)
		protected.POST("/roles/:id/tailor", resumeH.Tailor)
		protected.GET("/roles/:id/tailored-resume", resumeH.GetTailored)
		protected.GET("/settings/ai", aiH.GetSettings)
		protected.PUT("/settings/ai", aiH.UpdateSettings)
		protected.POST("/onboarding/complete", onboarding.Complete)
	}

}
