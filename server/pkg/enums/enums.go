package enums

// BookStatus represents the overall state of a book
type BookStatus string

const (
	BookDraft      BookStatus = "draft"      // Empty book, no volumes
	BookProcessing BookStatus = "processing" // At least one volume is being processed
	BookCompleted  BookStatus = "completed"  // All volumes parsed and ready
	BookError      BookStatus = "error"      // Critical error in book processing
)

func (bs BookStatus) ToString() string {
	return string(bs)
}

func (bs BookStatus) IsValid() bool {
	switch bs {
	case BookDraft, BookProcessing, BookCompleted, BookError:
		return true
	default:
		return false
	}
}

// VolumeStatus represents the state of a volume through its lifecycle
type VolumeStatus string

const (
	VolumeCreated   VolumeStatus = "created"   // Volume record created, no file yet
	VolumeUploaded  VolumeStatus = "uploaded"  // File uploaded to storage
	VolumeParsing   VolumeStatus = "parsing"   // Extracting structure (chapters/sections)
	VolumeParsed    VolumeStatus = "parsed"    // Structure extracted, stored in DB
	VolumeEnhancing VolumeStatus = "enhancing" // LLM processing (summaries, scenes)
	VolumeCompleted VolumeStatus = "completed" // Fully processed and ready
	VolumeError     VolumeStatus = "error"     // Error during processing
)

func (vs VolumeStatus) ToString() string {
	return string(vs)
}

func (vs VolumeStatus) IsValid() bool {
	switch vs {
	case VolumeCreated, VolumeUploaded, VolumeParsing, VolumeParsed,
		VolumeEnhancing, VolumeCompleted, VolumeError:
		return true
	default:
		return false
	}
}

// Can transition from current status to next status?
func (vs VolumeStatus) CanTransitionTo(next VolumeStatus) bool {
	validTransitions := map[VolumeStatus][]VolumeStatus{
		VolumeCreated:   {VolumeUploaded, VolumeError},
		VolumeUploaded:  {VolumeParsing, VolumeError},
		VolumeParsing:   {VolumeParsed, VolumeError},
		VolumeParsed:    {VolumeEnhancing, VolumeCompleted, VolumeError},
		VolumeEnhancing: {VolumeCompleted, VolumeError},
		VolumeCompleted: {VolumeParsing, VolumeEnhancing}, // Allow re-processing
		VolumeError:     {VolumeParsing, VolumeEnhancing}, // Allow retry
	}

	allowed, exists := validTransitions[vs]
	if !exists {
		return false
	}

	for _, s := range allowed {
		if s == next {
			return true
		}
	}
	return false
}

// ChapterStatus represents the processing state of a chapter
type ChapterStatus string

const (
	ChapterParsed    ChapterStatus = "parsed"    // Extracted from volume
	ChapterEnhancing ChapterStatus = "enhancing" // LLM processing
	ChapterCompleted ChapterStatus = "completed" // Fully processed
	ChapterError     ChapterStatus = "error"     // Error during processing
)

func (cs ChapterStatus) ToString() string {
	return string(cs)
}

// SectionStatus represents the processing state of a section
type SectionStatus string

const (
	SectionParsed    SectionStatus = "parsed"    // Extracted from chapter
	SectionEnhancing SectionStatus = "enhancing" // LLM processing
	SectionCompleted SectionStatus = "completed" // Fully processed
	SectionError     SectionStatus = "error"     // Error during processing
)

func (ss SectionStatus) ToString() string {
	return string(ss)
}

// ParsingMethod indicates how content was parsed
type ParsingMethod string

const (
	ParseMethodEPUBMetadata ParsingMethod = "epub_metadata" // From EPUB OPF/metadata
	ParseMethodEPUBContent  ParsingMethod = "epub_content"  // From EPUB HTML structure
	ParseMethodPDFMetadata  ParsingMethod = "pdf_metadata"  // From PDF metadata fields
	ParseMethodPDFLayout    ParsingMethod = "pdf_layout"    // From PDF text layout
	ParseMethodTextPattern  ParsingMethod = "text_pattern"  // Pattern matching in plain text
	ParseMethodLLMInference ParsingMethod = "llm_inference" // LLM-based inference
	ParseMethodManual       ParsingMethod = "manual"        // User-provided
)

func (pm ParsingMethod) ToString() string {
	return string(pm)
}

// JobType for the processing queue
type JobType string

const (
	JobTypeParse   JobType = "parse"   // Structure extraction
	JobTypeEnhance JobType = "enhance" // LLM enhancement
	JobTypeAudio   JobType = "audio"   // TTS generation
	JobTypeImage   JobType = "image"   // Image generation
)

func (jt JobType) ToString() string {
	return string(jt)
}
