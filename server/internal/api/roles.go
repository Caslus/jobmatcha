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
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}
	filter := service.NewRuleFilter(cfg)

	// Fetch all roles (not hidden)
	roles, err := h.repos.Role.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	// Filter by max age if set
	maxDays := cfg.MaxDaysOld
	if maxDays > 0 {
		// +1 day buffer so a job posted 26h ago (shown as "1d") still qualifies
		cutoff := time.Now().AddDate(0, 0, -(maxDays + 1))
		var recent []model.Role
		for _, r := range roles {
			if r.PostedAt != nil && r.PostedAt.After(cutoff) {
				recent = append(recent, r)
			}
		}
		roles = recent
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

	// Sort by percent DESC, then score DESC, then posted_at DESC
	sort.Slice(visible, func(i, j int) bool {
		if visible[i].Percent != visible[j].Percent {
			return visible[i].Percent > visible[j].Percent
		}
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
	items := make([]model.RoleListItem, len(paginated))
	for i, sr := range paginated {
		items[i] = model.RoleListItem{
			ID:             sr.Role.ID,
			CompanyID:      sr.Role.CompanyID,
			CompanyName:    sr.Role.Company.Name,
			Title:          sr.Role.Title,
			Department:     sr.Role.Department,
			Location:       sr.Role.Location,
			PostedAt:       sr.Role.PostedAt,
			RelevanceScore: sr.Score,
			MatchPercent:   sr.Percent,
			IsHidden:       sr.Role.IsHidden,
			IsInterested:   sr.Role.IsInterested,
		}
	}

	// Get total unfiltered count
	totalAll, _ := h.repos.Role.CountAll()

	c.JSON(http.StatusOK, model.RoleListResponse{
		Data: items,
		Pagination: model.PaginationInfo{
			Total:      total,
			Page:       page,
			PerPage:    perPage,
			TotalPages: totalPages,
		},
		TotalAll: int(totalAll),
	})
}

// GET /api/roles/:id
func (h *RoleHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid role ID."})
		return
	}

	role, err := h.repos.Role.GetByID(uint(id))
	if err != nil || role == nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Role not found."})
		return
	}

	// Score this role
	cfg, err := h.repos.Config.Get()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}
	filter := service.NewRuleFilter(cfg)
	sr := filter.Evaluate(role)

	score := 0
	percent := 0
	var reasons []string
	var details *model.MatchDetails
	if sr != nil {
		score = sr.Score
		percent = sr.Percent
		reasons = sr.Reasons

		// Build match details
		includeScore := sr.IncludeScore

		// Recency factor
		recency := 1.0
		if role.PostedAt != nil {
			days := time.Since(*role.PostedAt).Hours() / 24
			recency = 1.0 - days/180.0
			if recency < 0.3 {
				recency = 0.3
			}
		} else {
			recency = 0.3
		}
		adjusted := float64(includeScore) * recency

		details = &model.MatchDetails{
			IncludeScore:  includeScore,
			BonusScore:    sr.BonusScore,
			TotalScore:    score,
			RecencyFactor: recency,
			AdjustedScore: adjusted,
			MatchedKw:     len(sr.Reasons),
			TotalKw:       len(filter.IncludeKeywords),
			Percent:       percent,
		}
	}

	c.JSON(http.StatusOK, model.RoleDetailResponse{
		ID:                role.ID,
		CompanyID:         role.CompanyID,
		CompanyName:       role.Company.Name,
		Title:             role.Title,
		Department:        role.Department,
		Location:          role.Location,
		Description:       role.Description,
		DescriptionFormat: role.DescriptionFormat,
		URL:               role.URL,
		PostedAt:          role.PostedAt,
		RelevanceScore:    score,
		MatchPercent:      percent,
		MatchReasons:      reasons,
		MatchDetails:      details,
		IsHidden:          role.IsHidden,
		IsInterested:      role.IsInterested,
	})
}

// PATCH /api/roles/:id
func (h *RoleHandler) Patch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid role ID."})
		return
	}

	var req struct {
		IsHidden     *bool `json:"is_hidden,omitempty"`
		IsInterested *bool `json:"is_interested,omitempty"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request."})
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
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "No fields to update."})
		return
	}

	if err := h.repos.Role.Patch(uint(id), updates); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
