package models

import (
	"gorm.io/gorm"
)

type Asset struct {
	gorm.Model

	OwnerType string // scene / character_version / summary / tts / misc
	OwnerID   uint   `gorm:"index"`
	FileURL   string
	FileType  string
}
