package llm

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"google.golang.org/genai"
)

type GeminiService struct {
	client *genai.Client
	model  string
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
	fmt.Printf("LLM Service initialized (Provider: Gemini, Model: %s)", s.model)
	return nil
}

func (s *GeminiService) HealthCheck() error {
	if s.client == nil {
		return errors.New("gemini client is not initialized")
	}
	// Simple probe to check connectivity (optional, keeps costs low)
	return nil
}

func (s *GeminiService) GenerateJSON(ctx context.Context, sysPrompt, userPrompt string, schema any) (string, error) {
	if s.client == nil {
		return "", errors.New("gemini client is not initialized")
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
