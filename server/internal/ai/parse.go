package ai

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/caslus/jobmatcha/internal/model"
)

// parseJSONResult parses the LLM's text response into a ParseResumeResult.
// The response may be wrapped in markdown code fences or be bare JSON.
func parseJSONResult(text string) (*ParseResumeResult, error) {
	cleaned := text

	// Strip markdown code fences if present
	if idx := strings.Index(cleaned, "```"); idx >= 0 {
		rest := cleaned[idx+3:]
		if langEnd := strings.Index(rest, "\n"); langEnd >= 0 {
			// Skip optional language identifier
			rest = rest[langEnd+1:]
		}
		if end := strings.LastIndex(rest, "```"); end >= 0 {
			rest = rest[:end]
		}
		cleaned = strings.TrimSpace(rest)
	}

	// Remove any leading/trailing whitespace
	cleaned = strings.TrimSpace(cleaned)

	var result ParseResumeResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		// Try to find JSON object in the text
		if braceStart := strings.Index(cleaned, "{"); braceStart >= 0 {
			if braceEnd := strings.LastIndex(cleaned, "}"); braceEnd > braceStart {
				cleaned = cleaned[braceStart : braceEnd+1]
				if err2 := json.Unmarshal([]byte(cleaned), &result); err2 != nil {
					return nil, fmt.Errorf("json parse after brace extraction: %w (original: %v)", err2, err)
				}
				return &result, nil
			}
		}
		return nil, fmt.Errorf("json parse: %w\nraw text: %.200s", err, text)
	}
	return &result, nil
}

type tailorResumeResult struct {
	Document model.ResumeDocument `json:"document"`
}

func parseTailorResumeResultJSON(text string) (*tailorResumeResult, error) {
	cleaned := strings.TrimSpace(text)
	if idx := strings.Index(cleaned, "```"); idx >= 0 {
		rest := cleaned[idx+3:]
		if langEnd := strings.Index(rest, "\n"); langEnd >= 0 {
			rest = rest[langEnd+1:]
		}
		if end := strings.LastIndex(rest, "```"); end >= 0 {
			rest = rest[:end]
		}
		cleaned = strings.TrimSpace(rest)
	}

	var result tailorResumeResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		if braceStart := strings.Index(cleaned, "{"); braceStart >= 0 {
			if braceEnd := strings.LastIndex(cleaned, "}"); braceEnd > braceStart {
				if retryErr := json.Unmarshal([]byte(cleaned[braceStart:braceEnd+1]), &result); retryErr == nil {
					return &result, nil
				}
			}
		}
		return nil, fmt.Errorf("json parse: %w", err)
	}
	return &result, nil
}

// extractAssistantText extracts the plain text string from a ChatAssistantMessageContent union.
// The content can be either a simple string or an array of content items (text, image, etc.).
func extractAssistantText(content components.ChatAssistantMessageContent) string {
	if content.Str != nil {
		return strings.TrimSpace(*content.Str)
	}
	if len(content.ArrayOfChatContentItems) > 0 {
		var parts []string
		for _, item := range content.ArrayOfChatContentItems {
			if item.ChatContentText != nil {
				parts = append(parts, item.ChatContentText.Text)
			}
		}
		return strings.TrimSpace(strings.Join(parts, "\n"))
	}
	return ""
}
