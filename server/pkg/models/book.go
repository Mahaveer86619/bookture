package models

import (
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
	VolumeNo    int
	Description string

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
