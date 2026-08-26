package service

import (
	"context"

	"github.com/caslus/jobmatcha/internal/ai"
	"github.com/caslus/jobmatcha/internal/model"
)

// AIClient is the small boundary around the external AI provider used by HTTP
// handlers and resume workflows.
type AIClient interface {
	ValidateKey(context.Context, string) (bool, int, error)
	ParseResume(context.Context, string, string) (*ai.ParseResumeResult, error)
	TailorResume(context.Context, string, model.ResumeDocument, string, string, string, string) (*model.ResumeDocument, error)
}

type openRouterClient struct{}

// NewAIClient returns the production OpenRouter-backed client.
func NewAIClient() AIClient { return openRouterClient{} }

func (openRouterClient) ValidateKey(ctx context.Context, key string) (bool, int, error) {
	return ai.ValidateKey(ctx, key)
}

func (openRouterClient) ParseResume(ctx context.Context, key, text string) (*ai.ParseResumeResult, error) {
	return ai.ParseResume(ctx, key, text)
}

func (openRouterClient) TailorResume(ctx context.Context, key string, document model.ResumeDocument, title, company, location, description string) (*model.ResumeDocument, error) {
	return ai.TailorResume(ctx, key, document, title, company, location, description)
}
