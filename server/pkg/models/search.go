package models

import (
	"gorm.io/gorm"
)

type EmbeddingTarget struct {
	gorm.Model

	TargetType string
	TargetID   uint `gorm:"index"`

	Embeddings []Embedding `gorm:"constraint:OnDelete:CASCADE;"`
}

type Embedding struct {
	gorm.Model

	EmbeddingTargetID uint   `gorm:"index"`
	Vector            []byte `gorm:"type:bytea"`
	ModelName         string
}
