package services

import (
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/enums"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/services/llm"
	"gorm.io/gorm"
)

type ParserService struct {
	db  *gorm.DB
	llm llm.LLMService
}

func NewParserService(llmService llm.LLMService) *ParserService {
	return &ParserService{
		db:  db.GetBooktureDB().DB,
		llm: llmService,
	}
}

// BookMetadata defines the expected LLM output
type BookMetadata struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

func (ps *ParserService) ParseVolumeMetadata(volumeID uint, reportProgress func(int)) {
	log.Printf("Starting metadata parsing for Volume %d", volumeID)
	reportProgress(5) // Started

	var volume models.Volume
	if err := ps.db.Preload("Book").First(&volume, volumeID).Error; err != nil {
		log.Printf("Error fetching volume %d: %v", volumeID, err)
		return
	}

	reportProgress(10) // reading file

	if _, err := os.Stat(volume.FilePath); os.IsNotExist(err) {
		log.Printf("File not found: %s", volume.FilePath)
		return
	}

	// 1. Extract Text based on file type
	var sampleText string
	var err error

	ext := strings.ToLower(filepath.Ext(volume.FilePath))
	if ext == ".epub" {
		sampleText, err = ps.extractEpubText(volume.FilePath)
	} else {
		// Default text reading
		fileContent, readErr := os.ReadFile(volume.FilePath)
		if readErr == nil {
			sampleText = string(fileContent)
		}
		err = readErr
	}

	if err != nil {
		log.Printf("Failed to read file: %v", err)
		return
	}

	reportProgress(100)

	if len(sampleText) > 3000 {
		sampleText = sampleText[:3000]
	}

	// 2. Call LLM
	reportProgress(50)
	sysPrompt := `You are a literary analyst. Analyze the provided book text segment.
	Extract the Title, Author, and a short Description.
	Return strictly a JSON object.`

	jsonResp, err := ps.llm.GenerateJSON(context.Background(), sysPrompt, sampleText, nil)
	if err != nil {
		log.Printf("LLM Generation failed: %v", err)
		return
	}

	reportProgress(80) // Processing Response

	var meta BookMetadata
	if err := json.Unmarshal([]byte(jsonResp), &meta); err != nil {
		log.Printf("Failed to unmarshal LLM response: %v", err)
	}

	// 3. Update Database
	if meta.Title != "" {
		volume.Title = meta.Title
	}
	volume.Status = enums.Completed.ToString()
	volume.Progress = 100 // Persist completion
	ps.db.Save(&volume)

	// Update Book Metadata if it is still a placeholder
	var book models.Book
	if err := ps.db.First(&book, volume.BookID).Error; err == nil {
		updated := false
		if (book.Title == "Untitled draft" || book.Title == "") && meta.Title != "" {
			book.Title = meta.Title
			updated = true
		}
		if (book.Author == "Unknown" || book.Author == "") && meta.Author != "" {
			book.Author = meta.Author
			updated = true
		}
		if (book.Description == "No description provided" || book.Description == "") && meta.Description != "" {
			book.Description = meta.Description
			updated = true
		}

		if updated {
			ps.db.Save(&book)
			log.Printf("Updated Book %d metadata from Volume %d", book.ID, volume.ID)
		}
	}
}

func (ps *ParserService) extractEpubText(path string) (string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return "", err
	}
	defer r.Close()

	var textBuilder strings.Builder
	re := regexp.MustCompile(`<[^>]*>`) // Simple tag stripper

	for _, f := range r.File {
		// Heuristic: Read XHTML/HTML files from the zip
		if strings.HasSuffix(f.Name, ".html") || strings.HasSuffix(f.Name, ".xhtml") {
			rc, err := f.Open()
			if err != nil {
				continue
			}
			content, _ := io.ReadAll(rc)
			rc.Close()

			// Simple strip tags
			clean := re.ReplaceAllString(string(content), " ")
			textBuilder.WriteString(clean + "\n")

			if textBuilder.Len() > 5000 { // Optimization: Stop after enough text
				break
			}
		}
	}
	return textBuilder.String(), nil
}
