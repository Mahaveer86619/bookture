package gen_image

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
	"golang.org/x/time/rate"
)

type HuggingFaceDiffusersService struct {
	apiUrl  string
	apiKey  string
	client  *http.Client
	limiter *rate.Limiter
}

func (s *HuggingFaceDiffusersService) Init() error {
	s.apiUrl = config.AppConst.HuggingFaceStableDifussionXLbaseV1 + config.AppConfig.IMAGE_MODEL
	s.apiKey = config.AppConfig.IMAGE_KEY
	s.client = &http.Client{Timeout: 120 * time.Second}
	s.limiter = rate.NewLimiter(2, 5)

	if s.apiKey == "" {
		return fmt.Errorf("HuggingFace API key is missing")
	}

	fmt.Printf("Gen Image Service initialized (Provider: Hugging Face Diffusers, Model: %s)\n", config.AppConfig.IMAGE_MODEL)
	return nil
}

func (s *HuggingFaceDiffusersService) HealthCheck() error {
	return nil
}

func (s *HuggingFaceDiffusersService) GenerateImage(prompt string) (string, error) {
	if err := s.limiter.Wait(context.Background()); err != nil {
		return "", fmt.Errorf("rate limiter error: %w", err)
	}

	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		// Prepare request payload with 'wait_for_model' to assist the backoff
		payload := map[string]interface{}{
			"inputs":  prompt,
			"options": map[string]bool{"wait_for_model": true},
		}

		requestBody, err := json.Marshal(payload)
		if err != nil {
			return "", fmt.Errorf("marshal failed: %w", err)
		}

		req, err := http.NewRequest("POST", s.apiUrl, bytes.NewBuffer(requestBody))
		if err != nil {
			return "", fmt.Errorf("request creation failed: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+s.apiKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := s.client.Do(req)
		if err != nil {
			return "", fmt.Errorf("network error: %w", err)
		}

		// Immediate read and close to prevent resource leaks in loop
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("read response failed: %w", err)
		}

		// Handle Success
		if resp.StatusCode == http.StatusOK {
			// Verification: If content-type is JSON, HF sent an error message instead of an image
			if resp.Header.Get("Content-Type") == "application/json" {
				return "", fmt.Errorf("API returned JSON instead of image: %s", string(body))
			}
			return base64.StdEncoding.EncodeToString(body), nil
		}

		// Handle Retriable Errors (503 Model Loading or 429 Rate Limit)
		if resp.StatusCode == http.StatusTooManyRequests {
			// HF sends: RateLimit: "...;r=0;t=23"
			if rl := resp.Header.Get("RateLimit"); rl != "" {
				if reset := parseResetSeconds(rl); reset > 0 {
					time.Sleep(time.Duration(reset) * time.Second)
					continue
				}
			}

			// Fallback
			time.Sleep(time.Duration(math.Pow(2, float64(i+1))) * time.Second)
			continue
		}

		// Handle Non-retriable Errors
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	return "", fmt.Errorf("failed after %d retries", maxRetries)
}

func parseResetSeconds(header string) int {
	// example: "api|resolvers";r=0;t=23
	parts := strings.Split(header, ";")
	for _, p := range parts {
		if strings.HasPrefix(strings.TrimSpace(p), "t=") {
			if v, err := strconv.Atoi(strings.TrimPrefix(strings.TrimSpace(p), "t=")); err == nil {
				return v
			}
		}
	}
	return 0
}
