package api

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/caslus/jobmatcha/internal/model"
	"github.com/caslus/jobmatcha/internal/repository"
	"github.com/caslus/jobmatcha/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
)

type AIHandler struct {
	cfgRepo  *repository.ConfigRepo
	aiClient service.AIClient
}

func NewAIHandler(cfgRepo *repository.ConfigRepo, aiClient service.AIClient) *AIHandler {
	return &AIHandler{cfgRepo: cfgRepo, aiClient: aiClient}
}

// POST /api/ai/validate-key
func (h *AIHandler) ValidateKey(c *gin.Context) {
	var req model.AIValidateKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request."})
		return
	}
	if req.Provider == "" || req.APIKey == "" {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Provider and api_key are required."})
		return
	}

	valid, count, err := h.aiClient.ValidateKey(c.Request.Context(), req.APIKey)
	if err != nil {
		slog.Warn("ai key validation failed", "error", err)
		c.JSON(http.StatusOK, model.AIValidateKeyResponse{
			Valid: false,
			Error: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.AIValidateKeyResponse{
		Valid:  valid,
		Models: count,
	})
}

// GET /api/settings/ai
func (h *AIHandler) GetSettings(c *gin.Context) {
	cfg, err := h.cfgRepo.Get(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	hasKey := cfg.AIApiKey != ""

	c.JSON(http.StatusOK, model.AIInfoResponse{
		Provider:     cfg.AIProvider,
		Enabled:      cfg.AIEnabled,
		HasAPIKey:    hasKey,
		UserName:     cfg.UserName,
		UserEmail:    cfg.UserEmail,
		UserLocation: cfg.UserLocation,
		UserLinkedin: cfg.UserLinkedin,
		UserGithub:   cfg.UserGithub,
	})
}

// PUT /api/settings/ai
func (h *AIHandler) UpdateSettings(c *gin.Context) {
	var req model.AIUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "Invalid request."})
		return
	}

	updates := map[string]interface{}{}
	if req.Provider != nil {
		updates["ai_provider"] = *req.Provider
	}
	if req.APIKey != nil {
		updates["ai_api_key"] = *req.APIKey
	}
	if req.Enabled != nil {
		updates["ai_enabled"] = *req.Enabled
	}
	if req.UserName != nil {
		updates["user_name"] = *req.UserName
	}
	if req.UserEmail != nil {
		updates["user_email"] = *req.UserEmail
	}
	if req.UserLocation != nil {
		updates["user_location"] = *req.UserLocation
	}
	if req.UserLinkedin != nil {
		updates["user_linkedin"] = *req.UserLinkedin
	}
	if req.UserGithub != nil {
		updates["user_github"] = *req.UserGithub
	}

	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Error: "No fields to update."})
		return
	}

	if err := h.cfgRepo.UpdateMap(c.Request.Context(), updates); err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Error: "Internal error."})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ---- helpers ----

func filepathExt(name string) string {
	if idx := strings.LastIndex(name, "."); idx >= 0 {
		return name[idx:]
	}
	return ""
}

// extractPDFText extracts text from a raw PDF byte slice using the ledongthuc/pdf library.
// Also extracts link URLs from PDF annotations by scanning raw bytes.
func extractPDFText(data []byte) (string, error) {
	rs := bytes.NewReader(data)
	reader, err := pdf.NewReader(rs, int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("opening pdf: %w", err)
	}

	textReader, err := reader.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extracting pdf text: %w", err)
	}

	text, err := io.ReadAll(textReader)
	if err != nil {
		return "", fmt.Errorf("reading pdf text: %w", err)
	}
	plain := strings.TrimSpace(string(text))
	if plain == "" {
		return "", fmt.Errorf("no text content found in PDF")
	}

	// Extract URLs from raw PDF bytes (link annotations in PDF use /URI(...))
	urls := extractPDFURLs(data)
	if len(urls) > 0 {
		plain += "\n\n" + strings.Join(urls, "\n")
	}

	return plain, nil
}

// extractPDFURLs scans raw PDF bytes for /URI(...) patterns and extracts the URLs.
// PDF annotations look like: /URI(mailto:...) or /URI(https://...)
// Some PDFs have double /URI/URI(...) — handle both.
func extractPDFURLs(data []byte) []string {
	content := string(data)
	var urls []string
	seen := map[string]bool{}

	pos := 0
	for pos < len(content) {
		// Find the next /URI or /URI( marker
		idx := strings.Index(content[pos:], "/URI(")
		if idx < 0 {
			// Try with space: /URI (
			idx2 := strings.Index(content[pos:], "/URI (")
			if idx2 < 0 {
				// Try bare /URI and scan forward for (
				idx3 := strings.Index(content[pos:], "/URI")
				if idx3 < 0 {
					break
				}
				idx = idx3 + 4
				// Scan forward for (
				rest := content[pos+idx:]
				parenIdx := strings.Index(rest, "(")
				if parenIdx < 0 {
					break
				}
				// Check if there's another /URI between here and the paren
				nextURI := strings.Index(rest, "/URI")
				if nextURI >= 0 && nextURI < parenIdx {
					// Double /URI/URI( — skip to the second one
					idx = pos + idx + nextURI + 4
				} else {
					idx = pos + idx + parenIdx
				}
			} else {
				idx = pos + idx2 + 5
			}
		} else {
			idx = pos + idx + 5
		}

		if idx >= len(content) {
			break
		}

		// idx now points to the character after '('
		start := idx
		end := start
		depth := 1
		for end < len(content) && depth > 0 {
			if content[end] == '\\' {
				end += 2
				continue
			}
			if content[end] == '(' {
				depth++
			} else if content[end] == ')' {
				depth--
			}
			if depth > 0 {
				end++
			}
		}
		if depth == 0 {
			url := content[start:end]
			if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "mailto:") {
				if !seen[url] {
					urls = append(urls, url)
					seen[url] = true
				}
			}
		}
		pos = end + 1
	}

	return urls
}

// decodePDFHexString attempts to decode a PDF hex string (FEFF BOM = UTF-16BE).
func decodePDFHexString(hex string) string {
	hex = strings.Map(func(r rune) rune {
		if r == ' ' || r == '\n' || r == '\r' || r == '\t' {
			return -1
		}
		return r
	}, hex)

	if len(hex)%2 != 0 {
		hex = "0" + hex
	}

	if len(hex) >= 4 && (hex[:4] == "FEFF" || hex[:4] == "feff") {
		bytes := make([]byte, len(hex)/2)
		for i := 0; i < len(hex); i += 2 {
			var b byte
			fmt.Sscanf(hex[i:i+2], "%02x", &b)
			bytes[i/2] = b
		}
		runes := make([]rune, 0, len(bytes)/2)
		for i := 0; i+1 < len(bytes); i += 2 {
			r := rune(bytes[i])<<8 | rune(bytes[i+1])
			if r >= 0x20 {
				runes = append(runes, r)
			}
		}
		return string(runes)
	}

	bytes := make([]byte, len(hex)/2)
	for i := 0; i < len(hex); i += 2 {
		var b byte
		fmt.Sscanf(hex[i:i+2], "%02x", &b)
		bytes[i/2] = b
	}
	return string(bytes)
}
