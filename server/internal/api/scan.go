package api

import (
	"net/http"
	"strconv"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

type ScanHandler struct {
	svc *service.ScannerService
}

func NewScanHandler(svc *service.ScannerService) *ScanHandler {
	return &ScanHandler{svc: svc}
}

// POST /api/scan — starts a background scan, returns scan ID immediately.
func (h *ScanHandler) Start(c *gin.Context) {
	jobID, err := h.svc.StartScan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "failed to start scan"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"scan_id": jobID, "status": "pending"})
}

// GET /api/scan/:id — returns a scan job by ID.
func (h *ScanHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid scan id"})
		return
	}

	job, err := h.svc.GetJob(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "scan not found"})
		return
	}

	c.JSON(http.StatusOK, job)
}

// GET /api/scan/latest — returns the most recent scan job.
func (h *ScanHandler) GetLatest(c *gin.Context) {
	job, err := h.svc.GetLatestJob()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	if job == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "no scans yet"})
		return
	}

	c.JSON(http.StatusOK, job)
}