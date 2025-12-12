package models

import (
	"gorm.io/gorm"
)

type Character struct {
	gorm.Model

	BookID                   uint `gorm:"index"`
	Name                     string
	Alias                    string
	InitialDescription       string `gorm:"type:text"`
	Role                     string
	FirstAppearanceSectionID *uint `gorm:"index"`
	InitialImageURL          string

	Versions      []CharacterVersion      `gorm:"constraint:OnDelete:CASCADE;"`
	Relationships []CharacterRelationship `gorm:"constraint:OnDelete:CASCADE;foreignKey:CharacterAID"`
}

type CharacterVersion struct {
	gorm.Model

	CharacterID          uint  `gorm:"index"`
	SectionID            *uint `gorm:"index"`
	VersionNumber        int
	Description          string `gorm:"type:text"`
	AppearanceNotes      string `gorm:"type:text"`
	TitleOrRole          string
	EmotionalState       string
	Abilities            string
	RelationshipsSummary string `gorm:"type:text"`
	ImageURL             string
}

type CharacterRelationship struct {
	gorm.Model

	CharacterAID     uint `gorm:"index"`
	CharacterBID     uint `gorm:"index"`
	RelationshipType string
	SectionID        *uint  `gorm:"index"`
	Summary          string `gorm:"type:text"`
}
