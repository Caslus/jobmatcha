package api

import (
	"net/http"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	cfgRepo *repository.ConfigRepo
}

func NewSettingsHandler(cfgRepo *repository.ConfigRepo) *SettingsHandler {
	return &SettingsHandler{cfgRepo: cfgRepo}
}

// GET /api/settings
func (h *SettingsHandler) Get(c *gin.Context) {
	cfg, err := h.cfgRepo.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, model.SettingsResponse{
		IncludeKeywords:  cfg.IncludeKeywords,
		ExcludeKeywords:  cfg.ExcludeKeywords,
		LocationKeywords: cfg.LocationKeywords,
		WorkTypes:        cfg.WorkTypes,
		EmploymentType:   cfg.EmploymentType,
		MaxDaysOld:       cfg.MaxDaysOld,
	})
}

// PUT /api/settings
func (h *SettingsHandler) Update(c *gin.Context) {
	var req model.SettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request"})
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

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "no fields to update"})
		return
	}

	if err := h.cfgRepo.UpdateMap(updates); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}