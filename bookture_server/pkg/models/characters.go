package models

import "github.com/google/uuid"

type Character struct {
	Base
	BookID     uuid.UUID `gorm:"type:uuid;not null;index"`
	Name       string    `gorm:"not null"`
	BasePrompt string    // Core visual traits

	// Relations
	Snapshots []CharacterSnapshot `gorm:"foreignKey:CharacterID;constraint:OnDelete:CASCADE"`
}

type CharacterSnapshot struct {
	Base
	CharacterID uuid.UUID `gorm:"type:uuid;not null;index"`

	// Scope (A snapshot belongs to a volume OR a chapter)
	VolumeID  *uuid.UUID `gorm:"type:uuid"`
	ChapterID *uuid.UUID `gorm:"type:uuid"`

	VisualChanges  string // "Short hair now"
	EmotionalState string // "Angry"
	CurrentPrompt  string // The computed prompt for this state
}
