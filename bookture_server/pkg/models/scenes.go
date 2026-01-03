package models

import (
	"github.com/google/uuid"
	"github.com/pgvector/pgvector-go"
)

type Scene struct {
	Base

	ChapterID   uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_chapter_scene"`
	SceneIndex  int       `gorm:"not null;uniqueIndex:idx_chapter_scene"`
	ContentText string    `gorm:"not null"`

	// Intelligence & Summaries
	SummaryVisual    string
	SummaryNarrative string
	ImportanceScore  float64 `gorm:"default:0.0"`

	// Embedding for Semantic Search (pgvector)
	// GORM will map this to the VECTOR(1536) type
	Embedding pgvector.Vector `gorm:"type:vector(1536)"`

	// Multi-modal Assets
	ImageURL    string
	ImagePrompt string
	AudioURL    string
}
