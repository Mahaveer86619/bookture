package llm

import (
	"context"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
)

type LLMService interface {
	Init() error
	HealthCheck() error
	GenerateJSON(ctx context.Context, sysPrompt, userPrompt string, schema any) (string, error)
}

func NewLLMService() LLMService {
	cfg := config.AppConfig

	switch cfg.LLM_PROVIDER {
	case "gemini-api":
		return &GeminiService{}
	case "ollama":
		return &OllamaService{}
	// Case "s3": return &S3Storage{...}
	default:
		return &GeminiService{}
	}
}
