package api

import (
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, repos *repository.Repositories, db *gorm.DB) {
	auth := NewAuthHandler(repos.Config, db)

	// Public routes
	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	r.POST("/api/auth/login", auth.Login)
	r.POST("/api/auth/setup", auth.Setup)
	r.POST("/api/auth/logout", auth.Logout)
	r.GET("/api/auth/status", auth.Status)

	// Protected routes
	protected := r.Group("/api")
	protected.Use(Authenticated(repos.Config, db))
	{
		protected.POST("/auth/change-password", auth.ChangePassword)
	}
}