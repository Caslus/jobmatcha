package api

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
)

const maxUploadSize = 10 << 20 // 10 MB

type ResumeHandler struct {
	service *service.ResumeService
}

func NewResumeHandler(repos *repository.Repositories) *ResumeHandler {
	return &ResumeHandler{service: service.NewResumeService(repos)}
}

// POST /api/ai/parse-resume
// The extracted text is saved before AI parsing so an upload is never lost if
// the provider is unavailable or the parsing request fails.
func (h *ResumeHandler) ParseUpload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadSize)
	if err := c.Request.ParseMultipartForm(maxUploadSize); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "File too large or invalid form."})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "File is required."})
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to read file."})
		return
	}

	resumeText, err := extractResumeText(header.Filename, header.Header.Get("Content-Type"), data)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: err.Error()})
		return
	}

	resume, err := h.service.Save(c.Request.Context(), header.Filename, header.Header.Get("Content-Type"), resumeText)
	if err != nil {
		slog.Error("saving uploaded resume failed", "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to save resume."})
		return
	}

	parsed, err := h.service.Parse(c.Request.Context(), resume)
	if err != nil {
		if err == service.ErrAIKeyNotConfigured {
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Resume saved, but no AI API key is configured. Set it in settings first."})
			return
		}
		slog.Error("resume parse failed", "resume_id", resume.ID, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Resume saved, but AI parsing failed."})
		return
	}

	c.JSON(http.StatusOK, model.ParseResumeResponse{
		Name:                      parsed.Name,
		Email:                     parsed.Email,
		Location:                  parsed.Location,
		LinkedinURL:               parsed.LinkedinURL,
		GithubURL:                 parsed.GithubURL,
		SuggestedInclude:          parsed.IncludeKw,
		SuggestedExclude:          parsed.ExcludeKw,
		SuggestedWorkTypes:        parsed.WorkTypes,
		SuggestedLocationKeywords: parsed.LocationKw,
		Resume: model.ResumeInfoResponse{
			ID: resume.ID, Filename: resume.Filename, MediaType: resume.MediaType, CreatedAt: resume.CreatedAt,
		},
	})
}

// POST /api/roles/:id/tailor
func (h *ResumeHandler) Tailor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid role ID."})
		return
	}

	tailored, err := h.service.Tailor(c.Request.Context(), uint(id))
	if err != nil {
		switch {
		case err == service.ErrNoResume:
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Upload a resume before tailoring a role."})
		case err == service.ErrAIKeyNotConfigured:
			c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "No AI API key configured. Set it in settings first."})
		case err == service.ErrRoleNotFound:
			c.JSON(http.StatusNotFound, model.ErrorResponse{Error: "Role not found."})
		default:
			slog.Error("resume tailoring failed", "role_id", id, "error", err)
			c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to tailor resume. Please try again."})
		}
		return
	}

	c.JSON(http.StatusOK, tailoredResumeResponse(tailored))
}

// GET /api/roles/:id/tailored-resume
func (h *ResumeHandler) GetTailored(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid role ID."})
		return
	}
	tailored, err := h.service.GetTailored(c.Request.Context(), uint(id))
	if err != nil {
		slog.Error("loading tailored resume failed", "role_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Failed to load tailored resume."})
		return
	}
	if tailored == nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	c.JSON(http.StatusOK, tailoredResumeResponse(tailored))
}

func tailoredResumeResponse(tailored *model.TailoredResume) model.TailoredResumeResponse {
	return model.TailoredResumeResponse{
		ID: tailored.ID, ResumeID: tailored.ResumeID, RoleID: tailored.RoleID,
		Document: tailored.Document, CreatedAt: tailored.CreatedAt, UpdatedAt: tailored.UpdatedAt,
	}
}

func extractResumeText(filename, contentType string, data []byte) (string, error) {
	ext := strings.ToLower(filepathExt(filename))
	var resumeText string
	switch {
	case ext == ".md" || ext == ".txt" || contentType == "text/markdown" || contentType == "text/plain":
		resumeText = string(data)
	case ext == ".pdf" || contentType == "application/pdf":
		text, err := extractPDFText(data)
		if err != nil {
			return "", fmt.Errorf("failed to extract text from PDF: %w", err)
		}
		resumeText = text
	default:
		return "", fmt.Errorf("unsupported file type: %s. Accepted: .md, .txt, .pdf", ext)
	}
	if strings.TrimSpace(resumeText) == "" {
		return "", fmt.Errorf("no text content found in file")
	}
	return resumeText, nil
}
