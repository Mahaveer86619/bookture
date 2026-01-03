package models

import (
	"github.com/google/uuid"
)

type Job struct {
	Base

	EntityID   uuid.UUID `gorm:"type:uuid;not null"`
	EntityType string    `gorm:"not null"` // book, scene, character
	TaskType   string    `gorm:"not null"` // summarize, generate_image
	Status     string    `gorm:"default:'pending'"`
	RetryCount int       `gorm:"default:0"`
	ErrorLog   string
}

func (Job) TableName() string {
	return "job_queue"
}
