package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
	"google.golang.org/genai"
)

func (eps *ParserService) generateScenesForChapterWithRetry(chapter models.Chapter) ([]views.GeneratedScene, error) {
	var lastErr error

	for attempt := 1; attempt <= eps.maxRetries; attempt++ {
		scenes, err := eps.generateScenesForChapter(chapter)
		if err == nil {
			return scenes, nil
		}

		lastErr = err

		sleepDuration := eps.retryDelay * time.Duration(attempt)
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "RESOURCE_EXHAUSTED") {
			log.Printf("Rate limit hit. Waiting 60s before retry...")
			sleepDuration = 60 * time.Second
		}

		if attempt < eps.maxRetries {
			time.Sleep(sleepDuration)
		}
	}

	return nil, fmt.Errorf("failed after %d attempts: %w", eps.maxRetries, lastErr)
}

func (eps *ParserService) generateScenesForChapter(chapter models.Chapter) ([]views.GeneratedScene, error) {
	// Build context from all sections
	var textBuilder strings.Builder
	textBuilder.WriteString(fmt.Sprintf("Chapter %d: %s\n\n", chapter.ChapterNo, chapter.Title))

	for _, section := range chapter.Sections {
		textBuilder.WriteString(fmt.Sprintf("Section %d:\n%s\n\n", section.SectionNo, section.CleanText))
	}

	chapterText := textBuilder.String()

	// Limit context size (max ~8000 words)
	words := strings.Fields(chapterText)
	if len(words) > 8000 {
		chapterText = strings.Join(words[:8000], " ") + "..."
	}

	// Define schema for LLM
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"scenes": {
				Type: genai.TypeArray,
				Items: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"section_number": {
							Type:        genai.TypeInteger,
							Description: "The section number this scene belongs to",
						},
						"summary": {
							Type:        genai.TypeString,
							Description: "A 2-3 sentence summary of what happens in this scene",
						},
						"importance_score": {
							Type:        genai.TypeNumber,
							Description: "How important this scene is to the story (0.0 to 1.0)",
						},
						"scene_type": {
							Type:        genai.TypeString,
							Description: "Type of scene: action, dialogue, exposition, climax, resolution",
						},
						"image_prompt": {
							Type:        genai.TypeString,
							Description: "A detailed visual prompt for image generation, describing the scene, characters, setting, mood, and style",
						},
						"characters": {
							Type: genai.TypeArray,
							Items: &genai.Schema{
								Type: genai.TypeString,
							},
							Description: "List of character names present in this scene",
						},
						"location": {
							Type:        genai.TypeString,
							Description: "Where the scene takes place",
						},
						"mood": {
							Type:        genai.TypeString,
							Description: "The emotional tone: tense, peaceful, joyful, dark, mysterious, etc.",
						},
					},
					Required: []string{"section_number", "summary", "importance_score", "scene_type", "image_prompt"},
				},
			},
		},
		Required: []string{"scenes"},
	}

	sysPrompt := `You are a narrative analyst for visual storytelling.
Your task is to analyze a chapter and identify key scenes for visual representation.

For each section, create a scene with:
1. A concise summary of the action/events
2. An importance score (0.0-1.0) - higher for pivotal moments
3. Scene type classification
4. A detailed image prompt that captures the visual essence

Image prompts should:
- Describe the scene composition, characters, setting, and mood
- Be specific about visual details (lighting, colors, atmosphere)
- Maintain consistency with the story's tone
- Be suitable for AI image generation (avoid text/dialogue in images)

Return a JSON object with an array of scenes.`

	userPrompt := fmt.Sprintf(`Analyze this chapter and generate scenes for visual storytelling:

%s

Create one scene per section. Focus on the most visually compelling moments.`, chapterText)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	jsonResp, err := eps.llm.GenerateJSON(ctx, sysPrompt, userPrompt, schema)
	if err != nil {
		return nil, fmt.Errorf("LLM generation failed: %w", err)
	}

	// Parse response
	var response views.SceneGenerationResponse
	if err := json.Unmarshal([]byte(jsonResp), &response); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	if len(response.Scenes) == 0 {
		return nil, fmt.Errorf("no scenes generated")
	}

	return response.Scenes, nil
}
