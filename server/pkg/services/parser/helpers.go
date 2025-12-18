package parser

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/Mahaveer86619/bookture/server/pkg/enums"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
)

type ParsedVolume struct {
	DetectedTitle       string
	DetectedAuthor      string
	DetectedDescription string
	ParseMethod         enums.ParsingMethod
	Chapters            []ParsedChapter
	WordCount           int
	Errors              []string
}

type ParsedChapter struct {
	ChapterNumber       int
	DetectedTitle       string
	DetectionMethod     string
	DetectionConfidence float64
	Sections            []ParsedSection
	WordCount           int
}

type ParsedSection struct {
	SectionNumber int
	RawText       string
	CleanText     string
	WordCount     int
	HasDialogue   bool
	HasAction     bool
}

type EPUBPackage struct {
	XMLName  xml.Name     `xml:"package"`
	Metadata EPUBMetadata `xml:"metadata"`
	Manifest EPUBManifest `xml:"manifest"`
	Spine    EPUBSpine    `xml:"spine"`
}

type EPUBMetadata struct {
	Title       []string `xml:"title"`
	Creator     []string `xml:"creator"`
	Description []string `xml:"description"`
	Language    string   `xml:"language"`
}

type EPUBManifest struct {
	Items []EPUBItem `xml:"item"`
}

type EPUBItem struct {
	ID        string `xml:"id,attr"`
	Href      string `xml:"href,attr"`
	MediaType string `xml:"media-type,attr"`
}

type EPUBSpine struct {
	ItemRefs []EPUBItemRef `xml:"itemref"`
}

type EPUBItemRef struct {
	IDRef string `xml:"idref,attr"`
}

type LLMMetadataResponse struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	Genre       string `json:"genre,omitempty"`
}

type LLMChapterResponse struct {
	Chapters []LLMChapter `json:"chapters"`
}

type LLMChapter struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
}

func (s *ParserService) updateVolumeStatus(volumeID uint, status enums.VolumeStatus, progress int) {
	s.db.Model(&models.Volume{}).Where("id = ?", volumeID).Updates(map[string]interface{}{
		"status":   status.ToString(),
		"progress": progress,
	})
}

func (s *ParserService) markVolumeError(volumeID uint, errorMsg string) {
	errors := []string{errorMsg}
	errorsJSON, _ := json.Marshal(errors)

	s.db.Model(&models.Volume{}).Where("id = ?", volumeID).Updates(map[string]interface{}{
		"status":         enums.VolumeError.ToString(),
		"parsing_errors": string(errorsJSON),
		"progress":       -1,
	})
}

func (s *ParserService) saveStructuredData(volume *models.Volume, parsed *ParsedVolume) error {
	log.Printf("Saving structured data for Volume %d: %d chapters", volume.ID, len(parsed.Chapters))

	// Save chapters and sections
	for _, parsedChapter := range parsed.Chapters {
		chapter := models.Chapter{
			VolumeID:            volume.ID,
			ChapterNo:           parsedChapter.ChapterNumber,
			Title:               parsedChapter.DetectedTitle,
			Status:              enums.ChapterParsed.ToString(),
			DetectionMethod:     parsedChapter.DetectionMethod,
			DetectionConfidence: parsedChapter.DetectionConfidence,
			WordCount:           parsedChapter.WordCount,
		}

		if err := s.db.Create(&chapter).Error; err != nil {
			return fmt.Errorf("failed to create chapter: %w", err)
		}

		// Save sections
		for _, parsedSection := range parsedChapter.Sections {
			section := models.Section{
				ChapterID:   chapter.ID,
				SectionNo:   parsedSection.SectionNumber,
				RawText:     parsedSection.RawText,
				CleanText:   parsedSection.CleanText,
				Status:      enums.SectionParsed.ToString(),
				WordCount:   parsedSection.WordCount,
				HasDialogue: parsedSection.HasDialogue,
				HasAction:   parsedSection.HasAction,
			}

			if err := s.db.Create(&section).Error; err != nil {
				return fmt.Errorf("failed to create section: %w", err)
			}
		}
	}

	// Count sections
	var sectionCount int64
	s.db.Model(&models.Section{}).
		Joins("JOIN chapters ON sections.chapter_id = chapters.id").
		Where("chapters.volume_id = ?", volume.ID).
		Count(&sectionCount)

	// Update volume statistics
	volume.ChapterCount = len(parsed.Chapters)
	volume.SectionCount = int(sectionCount)
	volume.WordCount = parsed.WordCount
	volume.ParseMethod = parsed.ParseMethod.ToString()

	if len(parsed.Errors) > 0 {
		errorsJSON, _ := json.Marshal(parsed.Errors)
		volume.ParsingErrors = string(errorsJSON)
	}

	return s.db.Save(volume).Error
}

func (s *ParserService) updateMetadata(volume *models.Volume, parsed *ParsedVolume) {
	// Update volume title if detected
	if parsed.DetectedTitle != "" && (volume.Title == "" || strings.Contains(volume.Title, filepath.Ext(volume.FilePath))) {
		volume.Title = parsed.DetectedTitle
		s.db.Save(volume)
	}

	// Update book metadata if needed
	var book models.Book
	if err := s.db.First(&book, volume.BookID).Error; err == nil {
		updated := false

		// Update title
		if (book.Title == "Untitled draft" || book.Title == "" || book.Title == "Untitled") && parsed.DetectedTitle != "" {
			book.Title = parsed.DetectedTitle
			updated = true
		}

		// Update author
		if (book.Author == "Unknown" || book.Author == "") && parsed.DetectedAuthor != "" {
			book.Author = parsed.DetectedAuthor
			updated = true
		}

		// Update description
		if (book.Description == "" || book.Description == "No description provided") && parsed.DetectedDescription != "" {
			book.Description = parsed.DetectedDescription
			updated = true
		}

		if updated {
			s.db.Save(&book)
			log.Printf("Updated Book %d metadata from Volume %d", book.ID, volume.ID)
		}
	}
}
