package parser

import (
	"archive/zip"
	"context"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/enums"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"google.golang.org/genai"
)

func (s *ParserService) parseFileStructure(volume *models.Volume) (*ParsedVolume, error) {
	switch volume.FileFormat {
	case "epub":
		return s.parseEPUB(volume.FilePath)
	case "pdf":
		return s.parsePDF(volume.FilePath)
	case "txt":
		return s.parseText(volume.FilePath)
	default:
		return nil, fmt.Errorf("unsupported file format: %s", volume.FileFormat)
	}
}

// ============================================================================
// EPUB PARSING (RULE-BASED)
// ============================================================================

func (s *ParserService) parseEPUB(filePath string) (*ParsedVolume, error) {
	log.Printf("Parsing EPUB: %s", filePath)

	parsed := &ParsedVolume{
		ParseMethod: enums.ParseMethodEPUBContent,
		Chapters:    []ParsedChapter{},
		Errors:      []string{},
	}

	// Step 1: Try to extract metadata from content.opf
	metadata, err := s.extractEPUBMetadata(filePath)
	if err != nil {
		parsed.Errors = append(parsed.Errors, fmt.Sprintf("Failed to extract EPUB metadata: %v", err))
	} else {
		parsed.ParseMethod = enums.ParseMethodEPUBMetadata
		if len(metadata.Title) > 0 {
			parsed.DetectedTitle = metadata.Title[0]
		}
		if len(metadata.Creator) > 0 {
			parsed.DetectedAuthor = metadata.Creator[0]
		}
		if len(metadata.Description) > 0 {
			parsed.DetectedDescription = metadata.Description[0]
		}
	}

	// Step 2: Extract content from HTML/XHTML files
	content, err := s.extractEPUBContent(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract EPUB content: %w", err)
	}

	// Step 3: Detect chapters from content
	chapters := s.detectChaptersFromText(content)
	parsed.Chapters = chapters

	// Step 4: Calculate statistics
	parsed.WordCount = 0
	for _, ch := range parsed.Chapters {
		parsed.WordCount += ch.WordCount
	}

	log.Printf("EPUB parsing completed: %d chapters, %d words", len(parsed.Chapters), parsed.WordCount)
	return parsed, nil
}

func (s *ParserService) extractEPUBMetadata(path string) (*EPUBMetadata, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// Find content.opf file
	var opfFile *zip.File
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, ".opf") {
			opfFile = f
			break
		}
	}

	if opfFile == nil {
		return nil, errors.New("content.opf not found in EPUB")
	}

	// Read and parse OPF
	rc, err := opfFile.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	var pkg EPUBPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}

	return &pkg.Metadata, nil
}

func (s *ParserService) extractEPUBContent(filePath string) (string, error) {
	r, err := zip.OpenReader(filePath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	// 1. Build a map of all files by name for quick lookup
	zipFileMap := make(map[string]*zip.File)
	for _, f := range r.File {
		// Store both the original name and normalized versions
		zipFileMap[f.Name] = f
		// Also store without leading slash if present
		normalized := strings.TrimPrefix(f.Name, "/")
		zipFileMap[normalized] = f
	}

	// 2. Find and Parse OPF
	var opfFile *zip.File
	var opfPath string
	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".opf") {
			opfFile = f
			opfPath = f.Name
			break
		}
	}

	if opfFile == nil {
		return "", errors.New("content.opf not found in EPUB")
	}

	rc, err := opfFile.Open()
	if err != nil {
		return "", err
	}

	data, err := io.ReadAll(rc)
	rc.Close()
	if err != nil {
		return "", err
	}

	var pkg EPUBPackage
	if err := xml.Unmarshal(data, &pkg); err != nil {
		return "", err
	}

	// 3. Build manifest map with proper path resolution
	opfDir := filepath.Dir(opfPath)        // Use filepath for better cross-platform support
	manifestMap := make(map[string]string) // ID -> normalized path

	for _, item := range pkg.Manifest.Items {
		// Decode URL-encoded characters
		href := item.Href
		if decoded, err := url.QueryUnescape(href); err == nil {
			href = decoded
		}

		// Resolve relative to OPF directory
		var fullPath string
		if strings.HasPrefix(href, "/") {
			fullPath = strings.TrimPrefix(href, "/")
		} else if opfDir != "" && opfDir != "." {
			fullPath = filepath.Join(opfDir, href)
		} else {
			fullPath = href
		}

		// Normalize path separators for ZIP (always forward slashes)
		fullPath = filepath.ToSlash(fullPath)
		fullPath = strings.TrimPrefix(fullPath, "./")

		manifestMap[item.ID] = fullPath
	}

	// 4. Process spine in order
	var textBuilder strings.Builder
	htmlTagRegex := regexp.MustCompile(`<[^>]*>`)

	for _, itemRef := range pkg.Spine.ItemRefs {
		targetPath, exists := manifestMap[itemRef.IDRef]
		if !exists {
			log.Printf("Spine item %s not found in manifest", itemRef.IDRef)
			continue
		}

		// Try multiple path variations
		var contentFile *zip.File
		pathVariations := []string{
			targetPath,
			strings.TrimPrefix(targetPath, "/"),
			"/" + targetPath,
			strings.ReplaceAll(targetPath, "\\", "/"),
		}

		for _, pathVar := range pathVariations {
			if f, ok := zipFileMap[pathVar]; ok {
				contentFile = f
				break
			}
		}

		if contentFile == nil {
			log.Printf("Content file not found in ZIP: %s (tried %v)", targetPath, pathVariations)
			continue
		}

		// Read content
		frc, err := contentFile.Open()
		if err != nil {
			log.Printf("Failed to open %s: %v", targetPath, err)
			continue
		}
		content, err := io.ReadAll(frc)
		frc.Close()
		if err != nil {
			log.Printf("Failed to read %s: %v", targetPath, err)
			continue
		}

		text := string(content)

		// Improved HTML cleanup - preserve paragraph structure
		text = strings.ReplaceAll(text, "</p>", "\n\n")
		text = strings.ReplaceAll(text, "<br>", "\n")
		text = strings.ReplaceAll(text, "<br/>", "\n")
		text = strings.ReplaceAll(text, "<br />", "\n")
		text = strings.ReplaceAll(text, "</div>", "\n\n")
		text = strings.ReplaceAll(text, "</h1>", "\n\n")
		text = strings.ReplaceAll(text, "</h2>", "\n\n")
		text = strings.ReplaceAll(text, "</h3>", "\n\n")
		text = strings.ReplaceAll(text, "</section>", "\n\n")
		text = strings.ReplaceAll(text, "</article>", "\n\n")

		// Strip remaining tags
		text = htmlTagRegex.ReplaceAllString(text, "")

		// Decode HTML entities
		text = html.UnescapeString(text)

		// Normalize excessive whitespace BUT preserve line breaks
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			// Collase multiple spaces within a line
			lines[i] = regexp.MustCompile(`[^\S\n]+`).ReplaceAllString(line, " ")
			lines[i] = strings.TrimSpace(lines[i])
		}
		text = strings.Join(lines, "\n")

		// Remove excessive blank lines (more than 2 consecutive)
		text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")

		textBuilder.WriteString(text)
		textBuilder.WriteString("\n\n")
	}

	result := strings.TrimSpace(textBuilder.String())
	if result == "" {
		return "", errors.New("no content extracted from EPUB")
	}

	return result, nil
}

// ============================================================================
// PDF PARSING (PLACEHOLDER)
// ============================================================================

func (s *ParserService) parsePDF(filePath string) (*ParsedVolume, error) {
	log.Printf("PDF parsing not yet implemented: %s", filePath)

	// TODO: Implement PDF parsing using a library like pdfcpu or unidoc
	// For now, return a basic structure

	return &ParsedVolume{
		DetectedTitle: "",
		ParseMethod:   enums.ParseMethodPDFLayout,
		Chapters: []ParsedChapter{
			{
				ChapterNumber:       1,
				DetectedTitle:       "Full Document",
				DetectionMethod:     "default",
				DetectionConfidence: 0.3,
				Sections: []ParsedSection{
					{
						SectionNumber: 1,
						RawText:       "PDF content extraction pending implementation",
						CleanText:     "PDF content extraction pending implementation",
						WordCount:     5,
					},
				},
				WordCount: 5,
			},
		},
		WordCount: 5,
		Errors:    []string{"PDF parsing not yet implemented"},
	}, nil
}

// ============================================================================
// TEXT PARSING (PLACEHOLDER)
// ============================================================================

func (s *ParserService) parseText(filePath string) (*ParsedVolume, error) {
	log.Printf("Parsing plain text file: %s", filePath)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read text file: %w", err)
	}

	text := string(content)
	chapters := s.detectChaptersFromText(text)

	parsed := &ParsedVolume{
		ParseMethod: enums.ParseMethodTextPattern,
		Chapters:    chapters,
		WordCount:   len(strings.Fields(text)),
	}

	return parsed, nil
}

// ============================================================================
// CHAPTER DETECTION (RULE-BASED)
// ============================================================================

func (s *ParserService) detectChaptersFromText(text string) []ParsedChapter {
	// Try play structure first
	if strings.Contains(strings.ToUpper(text), "ACT I") || strings.Contains(strings.ToUpper(text), "ACT 1") {
		return s.detectPlayStructure(text)
	}

	// Fall back to your existing chapter detection
	return s.detectRegularChapters(text)
}

func (s *ParserService) detectPlayStructure(text string) []ParsedChapter {
	actPattern := regexp.MustCompile(`(?i)^ACT\s+([IVXLCDM]+|\d+)\s*$`)
	scenePattern := regexp.MustCompile(`(?i)^SCENE\s+([IVXLCDM]+|\d+)[:\.]?\s*(.*?)$`)

	lines := strings.Split(text, "\n")
	var chapters []ParsedChapter
	var currentAct *ParsedChapter
	var currentSceneText strings.Builder
	var sceneNumber int

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect ACT boundaries
		if matches := actPattern.FindStringSubmatch(line); matches != nil {
			// Save previous act
			if currentAct != nil {
				if currentSceneText.Len() > 0 {
					currentAct.Sections = append(currentAct.Sections,
						s.createSection(sceneNumber, currentSceneText.String()))
				}
				chapters = append(chapters, *currentAct)
			}

			// Start new act
			actNum := len(chapters) + 1
			currentAct = &ParsedChapter{
				ChapterNumber:       actNum,
				DetectedTitle:       line,
				DetectionMethod:     "play_act_pattern",
				DetectionConfidence: 0.9,
			}
			sceneNumber = 0
			currentSceneText.Reset()
			continue
		}

		// Detect SCENE boundaries
		if matches := scenePattern.FindStringSubmatch(line); matches != nil {
			// Save previous scene
			if currentSceneText.Len() > 0 && currentAct != nil {
				currentAct.Sections = append(currentAct.Sections,
					s.createSection(sceneNumber, currentSceneText.String()))
				currentSceneText.Reset()
			}

			sceneNumber++
			currentSceneText.WriteString(line + "\n\n")
			continue
		}

		// Regular content
		if line != "" {
			currentSceneText.WriteString(line + "\n")
		} else {
			currentSceneText.WriteString("\n")
		}
	}

	// Save last act
	if currentAct != nil {
		if currentSceneText.Len() > 0 {
			currentAct.Sections = append(currentAct.Sections,
				s.createSection(sceneNumber, currentSceneText.String()))
		}

		// Calculate total word count
		currentAct.WordCount = 0
		for _, s := range currentAct.Sections {
			currentAct.WordCount += s.WordCount
		}

		chapters = append(chapters, *currentAct)
	}

	return chapters
}

func (s *ParserService) detectRegularChapters(text string) []ParsedChapter {
	chapterPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)^chapter\s+(\d+|one|two|three|four|five|six|seven|eight|nine|ten|[ivxlcdm]+)[:\s]+(.*?)$`),
		regexp.MustCompile(`(?i)^ch\.?\s+(\d+)[:\s]+(.*?)$`),
		regexp.MustCompile(`(?i)^(\d+)\.\s+(.*?)$`),
		regexp.MustCompile(`(?i)^part\s+(\d+|one|two|three)[:\s]+(.*?)$`),
		regexp.MustCompile(`(?i)^prologue[:\s]*(.*?)$`),
		regexp.MustCompile(`(?i)^epilogue[:\s]*(.*?)$`),
	}

	lines := strings.Split(text, "\n")
	var chapters []ParsedChapter
	var currentChapter *ParsedChapter
	currentText := strings.Builder{}

	chapterNum := 0

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			currentText.WriteString("\n")
			continue
		}

		// Check if this line is a chapter heading
		isChapterHeading := false
		var chapterTitle string

		for _, pattern := range chapterPatterns {
			if matches := pattern.FindStringSubmatch(line); matches != nil {
				isChapterHeading = true
				chapterNum++
				if len(matches) > 2 {
					chapterTitle = strings.TrimSpace(matches[2])
				} else {
					chapterTitle = fmt.Sprintf("Chapter %d", chapterNum)
				}
				break
			}
		}

		if isChapterHeading {
			// Save previous chapter
			if currentChapter != nil {
				sections := s.splitIntoSections(currentText.String())
				currentChapter.Sections = sections
				currentChapter.WordCount = 0
				for _, s := range sections {
					currentChapter.WordCount += s.WordCount
				}
				chapters = append(chapters, *currentChapter)
			}

			// Start new chapter
			currentChapter = &ParsedChapter{
				ChapterNumber:       chapterNum,
				DetectedTitle:       chapterTitle,
				DetectionMethod:     "regex_pattern",
				DetectionConfidence: 0.8,
			}
			currentText.Reset()
		} else {
			currentText.WriteString(line)
			currentText.WriteString("\n")
		}
	}

	// Save last chapter
	if currentChapter != nil {
		sections := s.splitIntoSections(currentText.String())
		currentChapter.Sections = sections
		currentChapter.WordCount = 0
		for _, s := range sections {
			currentChapter.WordCount += s.WordCount
		}
		chapters = append(chapters, *currentChapter)
	}

	// If no chapters detected, treat entire text as one chapter
	if len(chapters) == 0 {
		sections := s.splitIntoSections(text)
		wordCount := 0
		for _, s := range sections {
			wordCount += s.WordCount
		}
		chapters = append(chapters, ParsedChapter{
			ChapterNumber:       1,
			DetectedTitle:       "Full Text",
			DetectionMethod:     "default",
			DetectionConfidence: 0.5,
			Sections:            sections,
			WordCount:           wordCount,
		})
	}

	return chapters
}

// ============================================================================
// SECTION SPLITTING (RULE-BASED)
// ============================================================================

func (s *ParserService) splitIntoSections(text string) []ParsedSection {
	// Split by scene breaks or paragraph grous
	sceneBreakPattern := regexp.MustCompile(`\n\s*[\*\-_]{3,}\s*\n`)
	parts := sceneBreakPattern.Split(text, -1)

	var sections []ParsedSection
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Further split long parts into ~500-1000 word sections
		paragraphs := strings.Split(part, "\n\n")
		currentSection := strings.Builder{}
		currentWordCount := 0
		sectionNum := len(sections) + 1

		for _, para := range paragraphs {
			para = strings.TrimSpace(para)
			if para == "" {
				continue
			}

			words := len(strings.Fields(para))

			// If adding this paragraph exceeds 1000 words, save current section
			if currentWordCount > 0 && currentWordCount+words > 1000 {
				sections = append(sections, s.createSection(sectionNum, currentSection.String()))
				currentSection.Reset()
				currentWordCount = 0
				sectionNum++
			}

			currentSection.WriteString(para)
			currentSection.WriteString("\n\n")
			currentWordCount += words
		}

		// Save remaining content
		if currentSection.Len() > 0 {
			sections = append(sections, s.createSection(sectionNum, currentSection.String()))
		}
	}

	// Ensure at least one section
	if len(sections) == 0 {
		sections = append(sections, s.createSection(1, text))
	}

	return sections
}

func (s *ParserService) createSection(sectionNum int, text string) ParsedSection {
	cleanText := strings.TrimSpace(text)
	wordCount := len(strings.Fields(cleanText))

	// Simple heuristics for dialogue and action
	hasDialogue := strings.Contains(cleanText, "\"") || strings.Contains(cleanText, "") || strings.Contains(cleanText, "")
	hasAction := strings.Contains(cleanText, "!") ||
		regexp.MustCompile(`\b(ran|jumped|fought|attacked|screamed)\b`).MatchString(cleanText)

	return ParsedSection{
		SectionNumber: sectionNum,
		RawText:       text,
		CleanText:     cleanText,
		WordCount:     wordCount,
		HasDialogue:   hasDialogue,
		HasAction:     hasAction,
	}
}

// ============================================================================
// LLM METADATA ENHANCEMENT
// ============================================================================

func (s *ParserService) enhanceMetadataWithLLM(parsed *ParsedVolume, volume *models.Volume) {
	log.Printf("Enhancing metadata with LLM for Volume %d", volume.ID)

	// Get sample text (first 2000 words)
	sampleText := s.getSampleText(parsed, 2000)
	if sampleText == "" {
		log.Println("No content available for LLM enhancement")
		return
	}

	// Define JSON schema for Gemini
	schema := &genai.Schema{
		Type: genai.TypeObject,
		Properties: map[string]*genai.Schema{
			"title": {
				Type:        genai.TypeString,
				Description: "The title of the book",
			},
			"author": {
				Type:        genai.TypeString,
				Description: "The author's name",
			},
			"description": {
				Type:        genai.TypeString,
				Description: "A brief description or summary of the book (2-3 sentences)",
			},
			"genre": {
				Type:        genai.TypeString,
				Description: "The primary genre of the book",
			},
		},
		Required: []string{"title", "author", "description"},
	}

	sysPrompt := `You are a literary analyst. Analyze the provided book excerpt.
Extract the Title, Author, and a short Description (2-3 sentences).
If the title or author cannot be determined from the text, provide your best inference.
Return strictly a JSON object with the specified fields.`

	userPrompt := fmt.Sprintf("Analyze this book excerpt and extract metadata:\n\n%s", sampleText)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	jsonResp, err := s.llm.GenerateJSON(ctx, sysPrompt, userPrompt, schema)
	if err != nil {
		log.Printf("LLM metadata generation failed: %v", err)
		parsed.Errors = append(parsed.Errors, fmt.Sprintf("LLM enhancement failed: %v", err))
		return
	}

	// Parse LLM response
	var meta LLMMetadataResponse
	if err := json.Unmarshal([]byte(jsonResp), &meta); err != nil {
		log.Printf("Failed to unmarshal LLM response: %v", err)
		parsed.Errors = append(parsed.Errors, fmt.Sprintf("Failed to parse LLM response: %v", err))
		return
	}

	// Update parsed volume with LLM-enhanced metadata
	if meta.Title != "" && parsed.DetectedTitle == "" {
		parsed.DetectedTitle = meta.Title
		parsed.ParseMethod = enums.ParseMethodLLMInference
	}
	if meta.Author != "" && parsed.DetectedAuthor == "" {
		parsed.DetectedAuthor = meta.Author
	}
	if meta.Description != "" && parsed.DetectedDescription == "" {
		parsed.DetectedDescription = meta.Description
	}

	log.Printf("LLM metadata enhancement completed: Title=%s, Author=%s", meta.Title, meta.Author)
}

func (s *ParserService) getSampleText(parsed *ParsedVolume, maxWords int) string {
	var textBuilder strings.Builder
	wordCount := 0

	for _, chapter := range parsed.Chapters {
		for _, section := range chapter.Sections {
			words := strings.Fields(section.CleanText)
			remaining := maxWords - wordCount

			if remaining <= 0 {
				break
			}

			if len(words) <= remaining {
				textBuilder.WriteString(section.CleanText)
				textBuilder.WriteString("\n\n")
				wordCount += len(words)
			} else {
				// Take only what we need
				textBuilder.WriteString(strings.Join(words[:remaining], " "))
				wordCount = maxWords
				break
			}
		}

		if wordCount >= maxWords {
			break
		}
	}

	return textBuilder.String()
}
