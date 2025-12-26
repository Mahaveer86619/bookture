package llm

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"golang.org/x/time/rate"
	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
	model  string

	rpmLimiter *rate.Limiter

	mu        sync.Mutex
	dailyUsed int
	dailyMax  int
	resetAt   time.Time
}

func (s *GeminiService) Init() error {
	ctx := context.Background()
	apiKey := config.AppConfig.LLM_KEY
	if apiKey == "" {
		return errors.New("LLM_KEY is not set in configuration")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return fmt.Errorf("failed to create gemini client: %w", err)
	}

	s.client = client
	s.model = config.AppConfig.LLM_MODEL

	// Configure limits per model
	switch s.model {
	case "gemini-2.5-flash-lite":
		s.rpmLimiter = rate.NewLimiter(rate.Every(time.Minute/15), 3)
		s.dailyMax = 1000

	case "gemini-2.5-flash":
		s.rpmLimiter = rate.NewLimiter(rate.Every(time.Minute/10), 2)
		s.dailyMax = 200

	case "gemini-2.5-pro":
		s.rpmLimiter = rate.NewLimiter(rate.Every(time.Minute/5), 1)
		s.dailyMax = 50

	case "gemini-2.0-flash":
		s.rpmLimiter = rate.NewLimiter(rate.Every(time.Minute/10), 2)
		s.dailyMax = 0 // unlimited daily

	default:
		return fmt.Errorf("unsupported Gemini model: %s", s.model)
	}

	// Daily reset at midnight UTC
	s.resetAt = time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour)

	fmt.Printf("LLM Service initialized (Gemini, Model: %s)\n", s.model)
	return nil
}

func (s *GeminiService) HealthCheck() error {
	if s.client == nil {
		return errors.New("gemini client is not initialized")
	}
	// Simple probe to check connectivity (optional, keeps costs low)
	return nil
}

func (s *GeminiService) checkDailyLimit() error {
	if s.dailyMax == 0 {
		return nil
	}

	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	if now.After(s.resetAt) {
		s.dailyUsed = 0
		s.resetAt = now.Truncate(24 * time.Hour).Add(24 * time.Hour)
	}

	if s.dailyUsed >= s.dailyMax {
		return fmt.Errorf("daily Gemini quota exceeded (%d/%d)", s.dailyUsed, s.dailyMax)
	}

	s.dailyUsed++
	return nil
}

func (s *GeminiService) GenerateJSON(
	ctx context.Context,
	sysPrompt, userPrompt string,
	schema any,
) (string, error) {

	if s.client == nil {
		return "", errors.New("gemini client is not initialized")
	}

	// RPM limiter
	if err := s.rpmLimiter.Wait(ctx); err != nil {
		return "", fmt.Errorf("gemini RPM limit exceeded: %w", err)
	}

	// RPD limiter
	if err := s.checkDailyLimit(); err != nil {
		return "", err
	}

	// Configure for JSON output
	genConfig := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{Text: sysPrompt},
			},
		},
	}

	// Apply Schema if provided (Must be of type *genai.Schema)
	if schema != nil {
		if googleSchema, ok := schema.(*genai.Schema); ok {
			genConfig.ResponseSchema = googleSchema
		} else {
			log.Println("Warning: Invalid schema type passed to GeminiService, expected *genai.Schema")
		}
	}

	// Execute Request
	resp, err := s.client.Models.GenerateContent(ctx, s.model, genai.Text(userPrompt), genConfig)
	if err != nil {
		return "", fmt.Errorf("gemini generation error: %w", err)
	}

	// Extract Text
	if resp != nil && len(resp.Candidates) > 0 {
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				return part.Text, nil
			}
		}
	}

	return "", errors.New("empty response from llm")
}
