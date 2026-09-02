package api

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

type CompanyHandler struct {
	svc        *service.CompanyService
	discoverer service.CareerBoardDiscoverer
}

func NewCompanyHandler(svc *service.CompanyService, discoverer service.CareerBoardDiscoverer) *CompanyHandler {
	return &CompanyHandler{svc: svc, discoverer: discoverer}
}

func (h *CompanyHandler) List(c *gin.Context) {
	items, err := h.svc.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Could not load companies."})
		return
	}
	c.JSON(http.StatusOK, model.CompanyListResponse{Data: items})
}

func (h *CompanyHandler) UpdateActive(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid company ID."})
		return
	}
	var req model.CompanyActiveUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Active == nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "An active state is required."})
		return
	}
	item, err := h.svc.UpdateActive(uint(id), *req.Active)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Company not found."})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *CompanyHandler) UpdateActiveBulk(c *gin.Context) {
	var req model.CompanyBulkActiveUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Active == nil || len(req.CompanyIDs) == 0 || hasDuplicateIDs(req.CompanyIDs) {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Company IDs and an active state are required."})
		return
	}
	if err := h.svc.UpdateActiveBulk(req.CompanyIDs, *req.Active); err != nil {
		if errors.Is(err, service.ErrCompanyNotFound) {
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "One or more companies were not found."})
			return
		}
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Could not update companies."})
		return
	}
	h.List(c)
}

func (h *CompanyHandler) UpdateBoardActive(c *gin.Context) {
	companyID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || companyID == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid company ID."})
		return
	}
	boardID, err := strconv.ParseUint(c.Param("boardID"), 10, 64)
	if err != nil || boardID == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid board ID."})
		return
	}
	var req model.CareerBoardActiveUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Active == nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "An active state is required."})
		return
	}
	item, err := h.svc.UpdateBoardActive(uint(companyID), uint(boardID), *req.Active)
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Career board not found."})
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *CompanyHandler) Discover(c *gin.Context) {
	var req model.CareerBoardDiscoveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "A careers URL is required."})
		return
	}
	if h.discoverer == nil {
		c.JSON(http.StatusServiceUnavailable, model.ErrorResponse{Error: "Career-board discovery is unavailable."})
		return
	}
	result, err := h.discoverer.DiscoverCareerBoards(c.Request.Context(), req.CareersURL)
	if err != nil {
		slog.Warn("career board discovery failed", "error", err)
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Could not analyze the careers URL: " + err.Error()})
		return
	}
	if name, err := h.svc.SuggestedEmployerName(result.Candidates); err != nil {
		slog.Warn("career board employer suggestion failed", "error", err)
	} else if name != "" {
		result.EmployerNameSuggestion = name
	}
	c.JSON(http.StatusOK, result)
}

func (h *CompanyHandler) RegisterBoards(c *gin.Context) {
	var req model.CareerBoardRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid career-board selections."})
		return
	}
	if err := h.svc.RegisterBoards(req.Candidates); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Could not register career boards."})
		return
	}
	h.List(c)
}

func hasDuplicateIDs(ids []uint) bool {
	seen := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			return true
		}
		if _, ok := seen[id]; ok {
			return true
		}
		seen[id] = struct{}{}
	}
	return false
}
