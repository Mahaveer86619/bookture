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
	Status      string // draft / processing / completed

	Volumes    []Volume    `gorm:"constraint:OnDelete:CASCADE;"`
	Characters []Character `gorm:"constraint:OnDelete:CASCADE;"`
}

type Volume struct {
	gorm.Model

	BookID      uint `gorm:"index"`
	Title       string
	Index       int
	Description string

	Status   string // pending / uploaded / processing / completed / error
	Progress int
	FilePath string // Path to the file in storage (local/s3)
	Uploaded bool   // true if file is physically present
	ParsedAt *time.Time

	Book     Book
	Chapters []Chapter `gorm:"constraint:OnDelete:CASCADE;"`
}

type Chapter struct {
	gorm.Model

	VolumeID     uint `gorm:"index"`
	ChapterNo    int
	Title        string
	SummaryShort string

	Sections []Section `gorm:"constraint:OnDelete:CASCADE;"`
}

type Section struct {
	gorm.Model

	ChapterID     uint `gorm:"index"`
	SectionNo     int
	RawText       string `gorm:"type:text"`
	CleanText     string `gorm:"type:text"`
	TTSAudio      string
	HasMajorEvent bool

	Scenes []Scene `gorm:"constraint:OnDelete:CASCADE;"`
}
