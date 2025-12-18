package models

import (
	"time"

	"gorm.io/gorm"
)

type Book struct {
	gorm.Model

	LibraryID   uint `gorm:"index"`
	Title       string
	Author      string
	Description string
	CoverImage  string
	Status      string `gorm:"type:varchar(20)"` // Use enums.BookStatus

	// Metadata
	TotalVolumes     int
	CompletedVolumes int

	Volumes    []Volume    `gorm:"constraint:OnDelete:CASCADE;"`
	Characters []Character `gorm:"constraint:OnDelete:CASCADE;"`
}

type Volume struct {
	gorm.Model

	BookID      uint `gorm:"index"`
	Title       string
	Index       int
	Description string

	Status      string `gorm:"type:varchar(20)"` // Use enums.VolumeStatus
	Progress    int    `gorm:"default:0"`        // 0-100
	CompletedAt *time.Time

	// File information
	FilePath   string
	FileSize   int64
	FileFormat string // "epub", "pdf", "txt"
	Uploaded   bool
	UploadedAt *time.Time

	// Parsing metadata
	ParsedAt      *time.Time
	ParseMethod   string `gorm:"type:varchar(30)"` // Use enums.ParsingMethod
	ParsingErrors string `gorm:"type:text"`        // JSON array of errors

	// Statistics
	WordCount    int
	ChapterCount int
	SectionCount int

	// Enhancement tracking
	EnhancedAt          *time.Time
	EnhancementProgress int `gorm:"default:0"` // Separate from parsing progress

	Book     Book
	Chapters []Chapter `gorm:"constraint:OnDelete:CASCADE;"`
}

type Chapter struct {
	gorm.Model

	VolumeID     uint `gorm:"index"`
	ChapterNo    int
	Title        string
	SummaryShort string
	Status       string `gorm:"type:varchar(20)"` // Use enums.ChapterStatus

	// Parsing metadata
	DetectionMethod     string  // How was this chapter detected?
	DetectionConfidence float64 `gorm:"type:decimal(3,2)"` // 0.00 to 1.00

	// Position tracking
	StartPosition int // Byte offset or page number in original file
	EndPosition   int
	WordCount     int

	Sections []Section `gorm:"constraint:OnDelete:CASCADE;"`
}

type Section struct {
	gorm.Model

	ChapterID uint `gorm:"index"`
	SectionNo int
	Status    string `gorm:"type:varchar(20)"` // Use enums.SectionStatus

	// Content
	RawText   string `gorm:"type:text"` // Original text
	CleanText string `gorm:"type:text"` // Normalized text

	// Metadata
	WordCount     int
	HasDialogue   bool
	HasAction     bool
	HasMajorEvent bool

	Scenes []Scene `gorm:"constraint:OnDelete:CASCADE;"`
}
