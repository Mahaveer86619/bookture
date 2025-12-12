package models

import (
	"gorm.io/gorm"
)

type Scene struct {
	gorm.Model

	SectionID  uint `gorm:"index"`
	ImageURL   string
	Caption    string `gorm:"type:text"`
	AltText    string `gorm:"type:text"`
	StyleTag   string
	IsKeyframe bool

	AIPrompts        []AIPrompt        `gorm:"constraint:OnDelete:CASCADE;"`
	AIGenerationJobs []AIGenerationJob `gorm:"polymorphic:Target;polymorphicValue:scene;constraint:OnDelete:CASCADE;"`
}
