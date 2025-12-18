package parser

import (
	"fmt"
	"log"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/models"
)

func (s *ParserService) generateImagesForVolume(volumeID uint, reportProgress func(int)) error {
	log.Printf("Generating images for Volume %d", volumeID)

	// Fetch all scenes that need images
	var scenes []models.Scene
	if err := s.db.Joins("JOIN sections ON sections.id = scenes.section_id").
		Joins("JOIN chapters ON chapters.id = sections.chapter_id").
		Where("chapters.volume_id = ? AND scenes.image_url IS NULL", volumeID).
		Order("chapters.chapter_no ASC, sections.section_no ASC").
		Find(&scenes).Error; err != nil {
		return fmt.Errorf("failed to fetch scenes: %w", err)
	}

	if len(scenes) == 0 {
		log.Printf("No scenes found requiring images for volume %d", volumeID)
		return nil
	}

	totalScenes := len(scenes)
	baseProgress := 60  // Starting at 60%
	progressRange := 35 // 60% to 95%

	for i, scene := range scenes {
		// Generate image with retry
		imageBase64, err := s.generateImageWithRetry(scene.ImagePrompt)
		if err != nil {
			log.Printf("Failed to generate image for Scene %d: %v", scene.ID, err)
			// Continue with next scene instead of failing
			continue
		}

		// Update scene with image
		scene.ImageURL = imageBase64 // Store base64 for now
		if err := s.db.Save(&scene).Error; err != nil {
			log.Printf("Failed to save image for Scene %d: %v", scene.ID, err)
			continue
		}

		// Update progress
		currentProgress := baseProgress + ((i + 1) * progressRange / totalScenes)
		reportProgress(currentProgress)
	}

	return nil
}

func (s *ParserService) generateImageWithRetry(prompt string) (string, error) {
	var lastErr error

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		imageBase64, err := s.imageGen.GenerateImage(prompt)
		if err == nil {
			return imageBase64, nil
		}

		lastErr = err
		log.Printf("Image generation attempt %d/%d failed: %v", attempt, s.maxRetries, err)

		if attempt < s.maxRetries {
			// Exponential backoff
			time.Sleep(s.retryDelay * time.Duration(attempt))
		}
	}

	return "", fmt.Errorf("failed after %d attempts: %w", s.maxRetries, lastErr)
}
