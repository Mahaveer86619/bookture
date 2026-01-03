package views

import (
	"github.com/Mahaveer86619/bookture/pkg/models"
	"github.com/google/uuid"
)

type SceneResponse struct {
	ID               uuid.UUID `json:"id"`
	Index            int       `json:"index"`
	Content          string    `json:"content"`           // The text to read
	SummaryNarrative string    `json:"summary_narrative"` // For highlights mode
	ImageURL         string    `json:"image_url,omitempty"`
	AudioURL         string    `json:"audio_url,omitempty"`
	IsImportant      bool      `json:"is_important"` // Derived from ImportanceScore
}

func ToSceneResponse(s *models.Scene) SceneResponse {
	return SceneResponse{
		ID:               s.ID,
		Index:            s.SceneIndex,
		Content:          s.ContentText,
		SummaryNarrative: s.SummaryNarrative,
		ImageURL:         s.ImageURL,
		AudioURL:         s.AudioURL,
		// Example logic: Consider it important if score > 0.7
		IsImportant: s.ImportanceScore > 0.7,
	}
}
