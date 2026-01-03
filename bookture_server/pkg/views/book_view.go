package views

import (
	"time"

	"github.com/Mahaveer86619/bookture/pkg/models"
	"github.com/google/uuid"
)

type CreateBookRequest struct {
	Title        string `json:"title" binding:"required"`
	Author       string `json:"author"`
	SourceFormat string `json:"source_format" binding:"oneof=pdf epub txt"`
}

type BookResponse struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Status      string    `json:"status"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	VolumeCount int       `json:"volume_count"` // Computed field
}

func ToBookResponse(m *models.Book) BookResponse {
	return BookResponse{
		ID:          m.ID,
		Title:       m.Title,
		Author:      m.Author,
		Description: m.Description,
		Status:      m.Status,
		CreatedAt:   m.CreatedAt,
		VolumeCount: len(m.Volumes), // Assumes Volumes are preloaded
	}
}
