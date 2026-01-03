package views

import (
	"github.com/Mahaveer86619/bookture/pkg/models"
	"github.com/google/uuid"
)

type SearchRequest struct {
	Query string `json:"query" binding:"required"`
	TopK  int    `json:"top_k" binding:"min=1,max=20"`
}

type SearchResult struct {
	SceneID   uuid.UUID `json:"scene_id"`
	Content   string    `json:"content"`
	ImageURL  string    `json:"image_url"`
	Score     float32   `json:"similarity_score"` // From vector distance
	ChapterID uuid.UUID `json:"chapter_id"`
}

func ToSearchResult(s *models.Scene, distance float32) SearchResult {
	return SearchResult{
		SceneID:   s.ID,
		Content:   s.ContentText,
		ImageURL:  s.ImageURL,
		ChapterID: s.ChapterID,
		Score:     1 - distance, // Convert distance to similarity (rough approx)
	}
}
