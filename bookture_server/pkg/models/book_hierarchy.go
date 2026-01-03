package models

import "github.com/google/uuid"

type Book struct {
	Base

	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Title        string    `gorm:"not null"`
	Author       string
	Description  string
	SourceFormat string `gorm:"check:source_format IN ('pdf', 'epub', 'txt')"`
	Status       string `gorm:"default:'draft'"` // draft, processing, ready, error
	IsPublic     bool   `gorm:"default:false"`

	// Relations
	Volumes    []Volume    `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
	Characters []Character `gorm:"foreignKey:BookID;constraint:OnDelete:CASCADE"`
}

type Volume struct {
	Base

	BookID       uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_book_volume"`
	VolumeNumber int       `gorm:"not null;uniqueIndex:idx_book_volume"`
	Title        string
	Summary      string

	// Relations
	Chapters []Chapter `gorm:"foreignKey:VolumeID;constraint:OnDelete:CASCADE"`
}

type Chapter struct {
	Base

	VolumeID      uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_volume_chapter"`
	ChapterNumber int       `gorm:"not null;uniqueIndex:idx_volume_chapter"`
	Title         string
	RawText       string
	Summary       string

	// Relations
	Scenes []Scene `gorm:"foreignKey:ChapterID;constraint:OnDelete:CASCADE"`
}
