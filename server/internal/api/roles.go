package api

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	repos *repository.Repositories
}

func NewRoleHandler(repos *repository.Repositories) *RoleHandler {
	return &RoleHandler{repos: repos}
}

// RoleListItem is the response DTO for the role list.
type RoleListItem struct {
	ID             uint       `json:"id"`
	CompanyID      uint       `json:"company_id"`
	CompanyName    string     `json:"company_name"`
	Title          string     `json:"title"`
	Department     string     `json:"department"`
	Location       string     `json:"location"`
	PostedAt       *time.Time `json:"posted_at"`
	RelevanceScore int        `json:"relevance_score"`
	IsHidden       bool       `json:"is_hidden"`
	IsInterested   bool       `json:"is_interested"`
}

type RoleListResponse struct {
	Data       []RoleListItem `json:"data"`
	Pagination PaginationInfo `json:"pagination"`
}

type PaginationInfo struct {
	Total      int `json:"total"`
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	TotalPages int `json:"total_pages"`
}

type RoleDetailResponse struct {
	ID             uint       `json:"id"`
	CompanyID      uint       `json:"company_id"`
	CompanyName    string     `json:"company_name"`
	Title          string     `json:"title"`
	Department     string     `json:"department"`
	Location       string     `json:"location"`
	Description    string     `json:"description"`
	URL            string     `json:"url"`
	PostedAt       *time.Time `json:"posted_at"`
	RelevanceScore int        `json:"relevance_score"`
	IsHidden       bool       `json:"is_hidden"`
	IsInterested   bool       `json:"is_interested"`
}

// GET /api/roles
func (h *RoleHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("per_page", "25"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 25
	}

	// Load user config for relevance scoring
	cfg, err := h.repos.Config.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	filter := service.NewRuleFilter(cfg)

	// Fetch all roles (not hidden)
	roles, err := h.repos.Role.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	// Score and filter
	scored := filter.FilterRoles(roles)

	// Filter out hidden roles
	var visible []service.ScoredRole
	for _, sr := range scored {
		if !sr.Role.IsHidden {
			visible = append(visible, sr)
		}
	}

	// Sort by score DESC, then posted_at DESC
	sort.Slice(visible, func(i, j int) bool {
		if visible[i].Score != visible[j].Score {
			return visible[i].Score > visible[j].Score
		}
		// Equal scores: newer first
		ti := visible[i].Role.PostedAt
		tj := visible[j].Role.PostedAt
		if ti == nil && tj == nil {
			return visible[i].Role.ID < visible[j].Role.ID
		}
		if ti == nil {
			return false
		}
		if tj == nil {
			return true
		}
		return ti.After(*tj)
	})

	total := len(visible)
	totalPages := (total + perPage - 1) / perPage

	// Paginate
	start := (page - 1) * perPage
	if start >= total {
		start = total
	}
	end := start + perPage
	if end > total {
		end = total
	}
	paginated := visible[start:end]

	// Build response
	items := make([]RoleListItem, len(paginated))
	for i, sr := range paginated {
		items[i] = RoleListItem{
			ID:             sr.Role.ID,
			CompanyID:      sr.Role.CompanyID,
			CompanyName:    sr.Role.Company.Name,
			Title:          sr.Role.Title,
			Department:     sr.Role.Department,
			Location:       sr.Role.Location,
			PostedAt:       sr.Role.PostedAt,
			RelevanceScore: sr.Score,
			IsHidden:       sr.Role.IsHidden,
			IsInterested:   sr.Role.IsInterested,
		}
	}

	c.JSON(http.StatusOK, RoleListResponse{
		Data: items,
		Pagination: PaginationInfo{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
	})
}

// GET /api/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid role id"})
		return
	}

	role, err := h.repos.Role.GetByID(uint(id))
	if err != nil || role == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "role not found"})
		return
	}

	// Score this role
	cfg, err := h.repos.Config.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}
	filter := service.NewRuleFilter(cfg)
	sr := filter.Evaluate(role)

	score := 0
	if sr != nil {
		score = sr.Score
	}

	c.JSON(http.StatusOK, RoleDetailResponse{
		ID:             role.ID,
		CompanyID:      role.CompanyID,
		CompanyName:    role.Company.Name,
		Title:          role.Title,
		Department:     role.Department,
		Location:       role.Location,
		Description:    role.Description,
		URL:            role.URL,
		PostedAt:       role.PostedAt,
		RelevanceScore: score,
		IsHidden:       role.IsHidden,
		IsInterested:   role.IsInterested,
	})
}

// PATCH /api/roles/:id
func (h *RoleHandler) Patch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid role id"})
		return
	}

	var req struct {
		IsHidden     *bool `json:"is_hidden,omitempty"`
		IsInterested *bool `json:"is_interested,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "invalid request"})
		return
	}

	updates := map[string]interface{}{}
	if req.IsHidden != nil {
		updates["is_hidden"] = *req.IsHidden
	}
	if req.IsInterested != nil {
		updates["is_interested"] = *req.IsInterested
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "no fields to update"})
		return
	}

	if err := h.repos.Role.Patch(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
