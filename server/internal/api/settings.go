package api

import (
	"net/http"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
)

type SettingsHandler struct {
	cfgRepo   *repository.ConfigRepo
	scheduler service.Scheduler
}

func NewSettingsHandler(cfgRepo *repository.ConfigRepo, scheduler service.Scheduler) *SettingsHandler {
	return &SettingsHandler{cfgRepo: cfgRepo, scheduler: scheduler}
}

// GET /api/settings
func (h *SettingsHandler) Get(c *gin.Context) {
	cfg, err := h.cfgRepo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	c.JSON(http.StatusOK, model.SettingsResponse{
		IncludeKeywords:  cfg.IncludeKeywords,
		ExcludeKeywords:  cfg.ExcludeKeywords,
		LocationKeywords: cfg.LocationKeywords,
		WorkTypes:        cfg.WorkTypes,
		EmploymentType:   cfg.EmploymentType,
		MaxDaysOld:       cfg.MaxDaysOld,
		ScanEnabled:      cfg.ScanEnabled,
		ScanCronExpr:     cfg.ScanCronExpr,
		ScanTimezone:     cfg.ScanTimezone,
		NextScanAt:       h.scheduler.NextRun(),
	})
}

// PUT /api/settings
func (h *SettingsHandler) Update(c *gin.Context) {
	var req model.SettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request."})
		return
	}

	updates := map[string]interface{}{}
	if req.IncludeKeywords != nil {
		updates["include_keywords"] = *req.IncludeKeywords
	}
	if req.ExcludeKeywords != nil {
		updates["exclude_keywords"] = *req.ExcludeKeywords
	}
	if req.LocationKeywords != nil {
		updates["location_keywords"] = *req.LocationKeywords
	}
	if req.WorkTypes != nil {
		updates["work_types"] = *req.WorkTypes
	}
	if req.EmploymentType != nil {
		updates["employment_type"] = *req.EmploymentType
	}
	if req.MaxDaysOld != nil {
		updates["max_days_old"] = *req.MaxDaysOld
	}
	if req.ScanEnabled != nil {
		updates["scan_enabled"] = *req.ScanEnabled
	}
	if req.ScanCronExpr != nil {
		// Validate cron expression using the standard parser
		parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
		if _, err := parser.Parse(*req.ScanCronExpr); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid cron expression: " + err.Error()})
			return
		}
		updates["scan_cron_expr"] = *req.ScanCronExpr
	}
	if req.ScanTimezone != nil {
		// Validate the timezone before persisting
		if _, err := time.LoadLocation(*req.ScanTimezone); err != nil {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid timezone: " + err.Error()})
			return
		}
		updates["scan_timezone"] = *req.ScanTimezone
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "No fields to update."})
		return
	}

	if err := h.cfgRepo.UpdateMap(c.Request.Context(), updates); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	// Reload the scheduler so it picks up any scan config changes
	h.scheduler.ReloadSchedule()

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
