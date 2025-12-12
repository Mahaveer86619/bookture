package models

import (
	"gorm.io/gorm"
)

type Summary struct {
	gorm.Model

	TargetType string // section / chapter / book
	TargetID   uint   `gorm:"index"`
	Summary    string `gorm:"type:text"`
	ImageURL   string

	AIPrompts        []AIPrompt        `gorm:"constraint:OnDelete:CASCADE;"`
	AIGenerationJobs []AIGenerationJob `gorm:"polymorphic:Target;polymorphicValue:summary;constraint:OnDelete:CASCADE;"`
}
