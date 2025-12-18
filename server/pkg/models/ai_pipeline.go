package models

import (
	"gorm.io/gorm"
)

type AIPrompt struct {
	gorm.Model

	SceneID    *uint  `gorm:"index"`
	SummaryID  *uint  `gorm:"index"`
	PromptText string `gorm:"type:text"`
	ModelName  string
}

func (aiPrompt *AIPrompt) TableName() string {
	return "ai_prompts"
}

type AIGenerationJob struct {
	gorm.Model

	TargetType string // scene / summary / tts
	TargetID   uint   `gorm:"index"`
	JobType    string // image / text / tts

	PromptText string `gorm:"type:text"`
	Status     string // pending / generating / completed / failed

	RetryCount int `gorm:"default:0"`
	MaxRetries int `gorm:"default:3"`

	ErrorMsg  string `gorm:"type:text"`
	OutputURL string
}
