package api

import (
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine, repos *repository.Repositories) {
	api := r.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})
	}
}