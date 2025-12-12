package models

import (
	"gorm.io/gorm"
)

type Progress struct {
	gorm.Model

	UserID           uint `gorm:"index"`
	BookID           uint `gorm:"index"`
	CurrentChapterID uint `gorm:"index"`
	CurrentSectionID uint `gorm:"index"`
}

type Bookmark struct {
	gorm.Model

	UserID     uint   `gorm:"index"`
	TargetType string // section / scene
	TargetID   uint   `gorm:"index"`
	Note       string `gorm:"type:text"`
}

type Rating struct {
	gorm.Model

	UserID uint `gorm:"index"`
	BookID uint `gorm:"index"`
	Rating int
	Review string `gorm:"type:text"`
}

type Annotation struct {
	gorm.Model

	UserID          uint   `gorm:"index"`
	SectionID       uint   `gorm:"index"`
	HighlightedText string `gorm:"type:text"`
	Note            string `gorm:"type:text"`
}
