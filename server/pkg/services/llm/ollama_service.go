package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
)

type OllamaService struct {
	client *http.Client
}

func (s *OllamaService) Init() error {
	s.client = &http.Client{Timeout: 120 * time.Second}
	fmt.Printf("Local LLM Service initialized (Provider: Ollama, Model: %s)\n", config.AppConfig.LLM_MODEL)
	return nil
}

func (s *OllamaService) HealthCheck() error {
	resp, err := s.client.Get(config.AppConfig.LLM_HOST + "/api/tags")
	if err != nil {
		return fmt.Errorf("ollama not reachable: %w", err)
	}
	defer resp.Body.Close()
	return nil
}

func (s *OllamaService) GenerateJSON(ctx context.Context, sysPrompt, userPrompt string, schema any) (string, error) {
	// Ollama works best when instructions are part of the prompt
	fullPrompt := fmt.Sprintf("System: %s\n\nUser: %s", sysPrompt, userPrompt)

	payload := map[string]interface{}{
		"model":  config.AppConfig.LLM_MODEL,
		"prompt": fullPrompt,
		"stream": false,
		"format": "json",
	}

	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", config.AppConfig.LLM_HOST+"/api/generate", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	return result.Response, nil
}
