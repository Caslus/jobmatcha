package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	openrouter "github.com/OpenRouterTeam/go-sdk"
	"github.com/OpenRouterTeam/go-sdk/models/components"
	"github.com/OpenRouterTeam/go-sdk/optionalnullable"
	"github.com/caslus/jobmatcha/internal/model"
)

const (
	resumeParseModel    = "openai/gpt-4o-mini"
	requestTimeout      = 30 * time.Second
	resumeParseTimeout  = 90 * time.Second
	resumeTailorTimeout = 90 * time.Second
)

// newClient creates an OpenRouter SDK client with the given API key.
func newClient(apiKey string) *openrouter.OpenRouter {
	return openrouter.New(
		openrouter.WithSecurity(apiKey),
		openrouter.WithHTTPReferer("https://jobmatcha.app"),
		openrouter.WithXTitle("jobmatcha"),
	)
}

// TailorResume rewrites and prioritizes a candidate's existing experience for
// one role. The model returns a structured document that the frontend owns for
// preview, editing, and PDF export.
func TailorResume(ctx context.Context, apiKey string, resume model.ResumeDocument, roleTitle, companyName, roleLocation, roleDescription string) (*model.ResumeDocument, error) {
	client := newClient(apiKey)

	ctx, cancel := context.WithTimeout(ctx, resumeTailorTimeout)
	defer cancel()

	systemContent := components.CreateChatSystemMessageContentStr(
		"You are an expert resume editor. Tailor the supplied structured resume for the job. Return ONLY JSON matching the requested schema. " +
			"Preserve the document's structure exactly: header, section order, section headings, section kinds, item counts, entry counts, titles, organizations, locations, dates, contact details, and bullet counts must all remain unchanged. " +
			"Only make small factual wording improvements to summary, entry highlights, and section items where the source already supports job-relevant ATS keywords. The summary should explicitly position the candidate for the target role by using its role name or a truthful equivalent when supported by their experience; for example, a software engineer with reliability, observability, and incident-management experience may be described as a Software Reliability Engineer (SRE). This is positioning, not a claim that they held that job title. " +
			"Do not invent facts, skills, metrics, responsibilities, employers, titles, credentials, or keywords. Keep all text unchanged when no safe edit exists.",
	)
	systemMsg := components.CreateChatMessagesSystem(components.ChatSystemMessage{Content: systemContent})

	resumeJSON, err := json.Marshal(resume)
	if err != nil {
		return nil, fmt.Errorf("encoding resume document: %w", err)
	}
	userContent := components.CreateChatUserMessageContentStr(
		"JOB\nTitle: " + roleTitle + "\nCompany: " + companyName + "\nLocation: " + roleLocation + "\nDescription:\n" + roleDescription +
			"\n\nCANDIDATE RESUME JSON\n" + string(resumeJSON),
	)
	userMsg := components.CreateChatMessagesUser(components.ChatUserMessage{Content: userContent})

	schema := map[string]any{
		"type":                 "object",
		"properties":           map[string]any{"document": resumeDocumentSchema()},
		"required":             []any{"document"},
		"additionalProperties": false,
	}

	jsonSchema := components.ChatFormatJSONSchemaConfig{
		JSONSchema: components.ChatJSONSchemaConfig{
			Name:        "tailored_resume",
			Description: openrouter.String("A factual, tailored resume document"),
			Schema:      schema,
			Strict:      optionalnullable.From(openrouter.Bool(true)),
		},
		Type: components.ChatFormatJSONSchemaConfigTypeJSONSchema,
	}
	responseFormat := components.CreateResponseFormatJSONSchema(jsonSchema)
	modelName := resumeParseModel
	stream := false
	request := components.ChatRequest{Model: &modelName, Messages: []components.ChatMessages{systemMsg, userMsg}, ResponseFormat: &responseFormat, Stream: &stream}

	response, err := client.Chat.Send(ctx, request, nil)
	if err != nil {
		return nil, fmt.Errorf("resume tailor request: %w", err)
	}
	if response == nil || response.ChatResult == nil || len(response.ChatResult.GetChoices()) == 0 {
		return nil, fmt.Errorf("empty response from OpenRouter")
	}
	content, ok := response.ChatResult.GetChoices()[0].Message.GetContent().Get()
	if !ok || content == nil {
		return nil, fmt.Errorf("empty content in assistant response")
	}
	result, err := parseTailorResumeResultJSON(extractAssistantText(*content))
	if err != nil {
		return nil, fmt.Errorf("parsing tailored resume: %w", err)
	}
	document, err := mergeTailoredDocument(resume, result.Document)
	if err != nil {
		return nil, fmt.Errorf("validating tailored document: %w", err)
	}
	return &document, nil
}

// ValidateKey checks whether the given API key is valid by listing models.
// Returns valid=true, modelCount, and any error.
func ValidateKey(ctx context.Context, apiKey string) (bool, int, error) {
	client := newClient(apiKey)

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	resp, err := client.Models.List(ctx, nil)
	if err != nil {
		slog.Warn("openrouter key validation failed", "error", err)
		return false, 0, fmt.Errorf("validating key: %w", err)
	}
	if resp == nil {
		return false, 0, fmt.Errorf("empty response from OpenRouter")
	}

	models := resp.GetResult()
	count := 0
	if models.Data != nil {
		count = len(models.Data)
	}

	slog.Info("openrouter key validated", "models_available", count)
	return true, count, nil
}

// ParseResumeResult holds structured data extracted from a resume.
type ParseResumeResult struct {
	Name        string               `json:"name"`
	Email       string               `json:"email"`
	Location    string               `json:"location"`
	LinkedinURL string               `json:"linkedin_url,omitempty"`
	GithubURL   string               `json:"github_url,omitempty"`
	IncludeKw   []string             `json:"suggested_include"`
	ExcludeKw   []string             `json:"suggested_exclude"`
	WorkTypes   []string             `json:"suggested_work_types"`
	LocationKw  []string             `json:"suggested_location_keywords"`
	Document    model.ResumeDocument `json:"document"`
}

// ParseResume sends resume text to the LLM and extracts structured profile data.
func ParseResume(ctx context.Context, apiKey string, resumeText string) (*ParseResumeResult, error) {
	client := newClient(apiKey)

	ctx, cancel := context.WithTimeout(ctx, resumeParseTimeout)
	defer cancel()

	// Build the system prompt
	systemContent := components.CreateChatSystemMessageContentStr(
		"You are a resume parser. Extract structured information from the resume text provided by the user. " +
			"Respond ONLY with valid JSON matching the requested schema. " +
			"Extract the candidate's full name, email, location (city and country/state), LinkedIn URL (if present), GitHub URL (if present). " +
			"URLs are extracted separately from the PDF and appended at the end of the text — scan for them there. " +
			"For suggested_include: extract keywords that will MATCH job listings the candidate is interested in. " +
			"This MUST include EVERY technology, programming language, framework, tool, cloud platform, and methodology mentioned " +
			"in the resume. Split compound terms into individual keywords — for example 'Oracle SQL' becomes 'oracle','sql', " +
			"'Java Spring' becomes 'java','spring', 'React Native' becomes 'react','native'. " +
			"Also derive implicit job-title-relevant keywords from context: e.g. if the resume mentions " +
			"'Site Reliability' or 'Observability' or 'Incident Management' also include 'sre', 'reliability', 'incident-response'. " +
			"If the resume has a 'Languages' section with human languages (e.g. Portuguese, English, Japanese), " +
			"include them as keywords too (e.g. 'portuguese', 'english', 'japanese', 'bilingual'). " +
			"CRITICAL: Extract role/job title keywords from the candidate's experience and education. " +
			"For example if the resume says 'Software Engineer' and 'Frontend Developer', include 'software', 'engineer', " +
			"'software engineer', 'frontend', 'front-end', 'frontend developer', 'developer'. " +
			"Also include translations of these role titles in relevant languages based on the candidate's " +
			"Languages section — e.g. if they speak Portuguese include 'engenheiro', 'software', 'desenvolvedor'; " +
			"if they speak Japanese include 'ソフトウェア', 'エンジニア', 'フロントエンド', '開発者'; " +
			"if they speak Spanish include 'ingeniero', 'desarrollador'. " +
			"Make sure to include a wide range of related terms and translations. " +
			"Also include many other roles related to the candidate's experience and education, even if not explicitly mentioned. " +
			"For example, if the resume mentions Java the candidate is likely interested in 'java', 'programming', 'backend', etc. " +
			"ALL keywords must be lowercase. Be thorough — include all of them. " +
			"For suggested_exclude: extract keywords that would EXCLUDE job listings the candidate does NOT want to see. " +
			"These are typically role-level qualifiers, seniority labels, or job family names that indicate the listing is not relevant. " +
			"Examples: senior, lead, principal, staff, manager, director, head, specialist, recruiter, intern. " +
			"Also include translations of these exclusion keywords in every language the candidate speaks " +
			"(e.g. Portuguese: 'sênior', 'líder', 'principal', 'gerente', 'diretor', 'especialista', 'recrutador', 'estagiário'; " +
			"Japanese: 'シニア', 'リード', 'プリンシパル', 'テックリクルーター', 'マネージャー', 'ディレクター', 'スペシャリスト', 'トレジャリー', 'マーケティング'; " +
			"Spanish: 'senior', 'lider', 'principal', 'gerente', 'director', 'especialista', 'reclutador', 'practicante'). " +
			"Also exclude industries, fields, or role types unrelated to the candidate's profile. " +
			"Be thorough — if the resume mentions any seniority level above the candidate's current role, include it here. " +
			"ALL exclude keywords must be lowercase. " +
			"For suggested_work_types: choose from: internship, contract, part-time, full-time. " +
			"For suggested_location_keywords: extract location-related keywords from the resume. " +
			"Always include 'remote' and 'hybrid'. Also include the candidate's city, state/province, country, " +
			"and any region mentioned. For example if located in 'Curitiba, Brazil' include " +
			"'curitiba', 'paraná', 'brazil', 'brasil', 'remote', 'hybrid', 'latin america', 'south america'. " +
			"ALL location keywords must be lowercase. " +
			"For document: produce a layout-ready structured resume. Header holds name and ordered contact lines. Summary is the summary text or empty. Sections must preserve the source order. " +
			"Use kind 'experience' or 'education' for sections with entries, otherwise 'list'. Each experience/education entry must keep its title, organization, location, date_range, and ordered highlights separate. " +
			"For a skills or technologies list, items MUST be 2–6 labelled category rows, never one item per skill. Keep source order within each row; for example: 'Backend: Java, Spring, Go, REST APIs' and 'Cloud & Operations: AWS, Docker, Linux, Grafana'. " +
			"For certifications, languages, achievements, projects, and other lists, each meaningful source line is one ordered item. " +
			"Only repair PDF-extraction whitespace. Do not add, omit, reorder, or rewrite any resume content. " +
			"Always output valid JSON with no markdown formatting.",
	)

	systemMsg := components.CreateChatMessagesSystem(components.ChatSystemMessage{
		Content: systemContent,
	})

	// Build the user message with resume text
	userContent := components.CreateChatUserMessageContentStr(resumeText)
	userMsg := components.CreateChatMessagesUser(components.ChatUserMessage{
		Content: userContent,
	})

	// Response format: JSON schema for structured output
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name":                        map[string]any{"type": "string"},
			"email":                       map[string]any{"type": "string"},
			"location":                    map[string]any{"type": "string"},
			"linkedin_url":                map[string]any{"type": "string"},
			"github_url":                  map[string]any{"type": "string"},
			"suggested_include":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"suggested_exclude":           map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"suggested_work_types":        map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"suggested_location_keywords": map[string]any{"type": "array", "items": map[string]any{"type": "string"}},
			"document":                    resumeDocumentSchema(),
		},
		"required":             []any{"name", "email", "location", "linkedin_url", "github_url", "suggested_include", "suggested_exclude", "suggested_work_types", "suggested_location_keywords", "document"},
		"additionalProperties": false,
	}

	jsonSchema := components.ChatFormatJSONSchemaConfig{
		JSONSchema: components.ChatJSONSchemaConfig{
			Name:        "resume_parse",
			Description: openrouter.String("Extracted resume information"),
			Schema:      schema,
			Strict:      optionalnullable.From(openrouter.Bool(true)),
		},
		Type: components.ChatFormatJSONSchemaConfigTypeJSONSchema,
	}
	responseFormat := components.CreateResponseFormatJSONSchema(jsonSchema)

	model := resumeParseModel
	stream := false

	req := components.ChatRequest{
		Model:          &model,
		Messages:       []components.ChatMessages{systemMsg, userMsg},
		ResponseFormat: &responseFormat,
		Stream:         &stream,
	}

	resp, err := client.Chat.Send(ctx, req, nil)
	if err != nil {
		return nil, fmt.Errorf("resume parse request: %w", err)
	}
	if resp == nil || resp.ChatResult == nil {
		return nil, fmt.Errorf("empty response from OpenRouter")
	}

	choices := resp.ChatResult.GetChoices()
	if len(choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Extract the text content from the assistant message
	msg := choices[0].Message
	contentOpt := msg.GetContent()

	contentVal, contentOk := contentOpt.Get()
	if !contentOk || contentVal == nil {
		return nil, fmt.Errorf("empty content in assistant response")
	}

	contentStr := extractAssistantText(*contentVal)
	if contentStr == "" {
		return nil, fmt.Errorf("empty text content in assistant response")
	}

	// The content should be JSON. Parse it into our result struct.
	result, err := parseJSONResult(contentStr)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM response: %w", err)
	}

	slog.Info("resume parsed successfully",
		"name", result.Name,
		"email", result.Email,
		"include_kw_count", len(result.IncludeKw),
	)
	return result, nil
}
