package parser

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/enums"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"gorm.io/gorm"
)

func (s *ParserService) ProcessVolumeComplete(volumeID uint, reportProgress func(int)) {
	log.Printf("Starting complete processing pipeline for Volume %d", volumeID)

	// Phase 1: Parse Structure (0-30%)
	s.updateVolumeStatus(volumeID, enums.VolumeParsing, 0)
	reportProgress(5)

	var volume models.Volume
	if err := s.db.Preload("Book").First(&volume, volumeID).Error; err != nil {
		s.markVolumeError(volumeID, "Failed to fetch volume")
		return
	}

	parsed, err := s.parseFileStructure(&volume)
	if err != nil {
		s.markVolumeError(volumeID, err.Error())
		return
	}
	reportProgress(20)

	if parsed.DetectedTitle == "" || parsed.DetectedAuthor == "" {
		s.enhanceMetadataWithLLM(parsed, &volume)
	}
	reportProgress(25)

	if err := s.saveStructuredData(&volume, parsed); err != nil {
		s.markVolumeError(volumeID, err.Error())
		return
	}
	s.updateMetadata(&volume, parsed)
	reportProgress(30)

	// Phase 2: Generate Scenes with LLM (30-60%)
	s.updateVolumeStatus(volumeID, enums.VolumeEnhancing, 30)

	if err := s.generateScenesForVolume(volumeID, reportProgress); err != nil {
		log.Printf("Scene generation failed: %v", err)
		s.markVolumeError(volumeID, fmt.Sprintf("Scene generation failed: %v", err))
		return
	}
	reportProgress(60)

	// Phase 3: Generate Images (60-95%)
	if err := s.generateImagesForVolume(volumeID, reportProgress); err != nil {
		log.Printf("Image generation failed: %v", err)
		// Don't fail the entire process if images fail
		log.Printf("Continuing despite image generation errors")
	}
	reportProgress(95)

	// Mark as completed
	now := time.Now()
	volume.Status = enums.VolumeCompleted.ToString()
	volume.CompletedAt = &now
	volume.Progress = 100
	s.db.Save(&volume)

	reportProgress(100)
	log.Printf("Volume %d processing completed successfully", volumeID)
}

func (s *ParserService) generateScenesForVolume(volumeID uint, reportProgress func(int)) error {
	log.Printf("Generating scenes for Volume %d", volumeID)

	// Fetch all chapters with sections
	var chapters []models.Chapter
	if err := s.db.Where("volume_id = ?", volumeID).
		Order("chapter_no ASC").
		Preload("Sections", func(db *gorm.DB) *gorm.DB {
			return db.Order("section_no ASC")
		}).
		Find(&chapters).Error; err != nil {
		return fmt.Errorf("failed to fetch chapters: %w", err)
	}

	if len(chapters) == 0 {
		return fmt.Errorf("no chapters found for volume %d", volumeID)
	}

	totalSections := 0
	for _, ch := range chapters {
		totalSections += len(ch.Sections)
	}

	processedSections := 0
	baseProgress := 30  // Starting at 30%
	progressRange := 30 // 30% to 60%

	// Process each chapter
	for _, chapter := range chapters {
		if len(chapter.Sections) == 0 {
			continue
		}

		// Generate scenes for this chapter with retry
		scenes, err := s.generateScenesForChapterWithRetry(chapter)
		if err != nil {
			log.Printf("Failed to generate scenes for Chapter %d: %v", chapter.ID, err)
			// Continue with next chapter instead of failing entirely
			continue
		}

		// Save scenes to database
		for _, scene := range scenes {
			// Find corresponding section
			var section *models.Section
			for i := range chapter.Sections {
				if chapter.Sections[i].SectionNo == scene.SectionNumber {
					section = &chapter.Sections[i]
					break
				}
			}

			if section == nil {
				log.Printf("Section %d not found for scene", scene.SectionNumber)
				continue
			}

			// Create scene record
			sceneModel := models.Scene{
				SectionID:       section.ID,
				Summary:         scene.Summary,
				ImagePrompt:     scene.ImagePrompt,
				ImportanceScore: scene.ImportanceScore,
				SceneType:       scene.SceneType,
				Characters:      strings.Join(scene.Characters, ","),
				Location:        scene.Location,
				Mood:            scene.Mood,
				Status:          enums.SectionCompleted.ToString(),
			}

			if err := s.db.Create(&sceneModel).Error; err != nil {
				log.Printf("Failed to save scene: %v", err)
				continue
			}

			// Update section status
			section.Status = enums.SectionCompleted.ToString()
			s.db.Save(section)
		}

		// Update chapter status
		chapter.Status = enums.ChapterCompleted.ToString()
		s.db.Save(&chapter)

		// Update progress
		processedSections += len(chapter.Sections)
		currentProgress := baseProgress + (processedSections * progressRange / totalSections)
		reportProgress(currentProgress)
	}

	return nil
}

func (s *ParserService) RetrySceneGeneration(volumeID uint, reportProgress func(int)) error {
	log.Printf("Retrying scene generation for Volume %d", volumeID)

	// Delete existing scenes that failed or are incomplete
	if err := s.db.Exec(`
		DELETE FROM scenes 
		WHERE section_id IN (
			SELECT sections.id FROM sections
			JOIN chapters ON chapters.id = sections.chapter_id
			WHERE chapters.volume_id = ?
		)
	`, volumeID).Error; err != nil {
		return fmt.Errorf("failed to clear existing scenes: %w", err)
	}

	// Reset section and chapter status
	s.db.Exec(`
		UPDATE sections SET status = ?
		WHERE chapter_id IN (
			SELECT id FROM chapters WHERE volume_id = ?
		)
	`, enums.SectionParsed.ToString(), volumeID)

	s.db.Exec(`
		UPDATE chapters SET status = ?
		WHERE volume_id = ?
	`, enums.ChapterParsed.ToString(), volumeID)

	// Re-run scene generation
	return s.generateScenesForVolume(volumeID, reportProgress)
}

func (s *ParserService) RetryImageGeneration(volumeID uint, reportProgress func(int)) error {
	log.Printf("Retrying image generation for Volume %d", volumeID)

	// Clear existing images
	s.db.Exec(`
		UPDATE scenes SET image_url = NULL
		WHERE section_id IN (
			SELECT sections.id FROM sections
			JOIN chapters ON chapters.id = sections.chapter_id
			WHERE chapters.volume_id = ?
		)
	`, volumeID)

	// Re-run image generation
	return s.generateImagesForVolume(volumeID, reportProgress)
}
