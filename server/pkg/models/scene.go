package models

import (
	"gorm.io/gorm"
)

type Scene struct {
	gorm.Model

	SectionID uint `gorm:"index"`
	ImageURL  string
	Caption   string `gorm:"type:text"`

	Summary         string  `gorm:"type:text"`
	ImagePrompt     string  `gorm:"type:text"`
	ImportanceScore float64 `gorm:"type:decimal(3,2)"`
	SceneType       string  `gorm:"type:varchar(30)"`
	Characters      string  `gorm:"type:text"` // Comma separated names
	Location        string
	Mood            string
	Status          string `gorm:"type:varchar(20)"`

	IsKeyframe bool

	AIPrompts        []AIPrompt        `gorm:"constraint:OnDelete:CASCADE;"`
	AIGenerationJobs []AIGenerationJob `gorm:"polymorphic:Target;polymorphicValue:scene;constraint:OnDelete:CASCADE;"`
}
