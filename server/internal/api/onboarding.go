package api

import (
	"log/slog"
	"net/http"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

type OnboardingHandler struct {
	cfgRepo      *repository.ConfigRepo
	scanSvc      service.Scanner
	schedulerSvc service.Scheduler
}

func NewOnboardingHandler(cfgRepo *repository.ConfigRepo, scanSvc service.Scanner, schedulerSvc service.Scheduler) *OnboardingHandler {
	return &OnboardingHandler{cfgRepo: cfgRepo, scanSvc: scanSvc, schedulerSvc: schedulerSvc}
}

// POST /api/onboarding/complete — finalize onboarding, save profile, start first scan.
func (h *OnboardingHandler) Complete(c *gin.Context) {
	var req model.OnboardingCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request."})
		return
	}

	// Build the update map with all onboarding fields
	updates := map[string]interface{}{
		"user_name":         req.UserName,
		"user_email":        req.UserEmail,
		"user_location":     req.UserLocation,
		"user_linkedin":     req.UserLinkedin,
		"user_github":       req.UserGithub,
		"include_keywords":  req.IncludeKeywords,
		"exclude_keywords":  req.ExcludeKeywords,
		"location_keywords": req.LocationKeywords,
		"work_types":        req.WorkTypes,
		"max_days_old":      req.MaxDaysOld,
		"scan_enabled":      req.ScanEnabled,
		"scan_cron_expr":    req.ScanCronExpr,
		"scan_timezone":     req.ScanTimezone,
		"setup_complete":    true,
	}

	if err := h.cfgRepo.UpdateMap(c.Request.Context(), updates); err != nil {
		slog.Error("onboarding: failed to save config", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to save onboarding data."})
		return
	}

	// Reload scheduler to pick up new scan schedule
	h.schedulerSvc.ReloadSchedule()

	// Auto-start first scan in background
	jobID, err := h.scanSvc.StartScan()
	if err != nil {
		slog.Error("onboarding: failed to start first scan", "error", err)
		// Onboarding data saved; scan failure is non-fatal
		c.JSON(http.StatusOK, gin.H{"status": "ok", "setup_complete": true, "scan_started": false, "message": "Onboarding complete but first scan failed to start."})
		return
	}

	slog.Info("onboarding complete", "user", req.UserName, "scan_job", jobID)
	c.JSON(http.StatusOK, gin.H{
		"status":         "ok",
		"setup_complete": true,
		"scan_started":   true,
		"scan_id":        jobID,
	})
}
