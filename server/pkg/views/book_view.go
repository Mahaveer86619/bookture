package views

import (
	"errors"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/utils"
)

type BookView struct {
	ID          string    `json:"id"`
	LibraryID   string    `json:"library_id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	CoverImage  string    `json:"cover_image"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToBookView(b *models.Book) BookView {
	return BookView{
		ID:          utils.MaskID(b.ID),
		LibraryID:   utils.MaskID(b.LibraryID),
		Title:       b.Title,
		Author:      b.Author,
		Description: b.Description,
		CoverImage:  b.CoverImage,
		Status:      b.Status,
		CreatedAt:   b.CreatedAt,
		UpdatedAt:   b.UpdatedAt,
	}
}

func ToBookViews(books []models.Book) []BookView {
	views := make([]BookView, len(books))
	for i, b := range books {
		views[i] = ToBookView(&b)
	}
	return views
}

// CreateBookRequest is used when initializing a book upload or creating a draft
type CreateBookRequest struct {
	LibraryID   string `json:"library_id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
}

func (r CreateBookRequest) Valid() error {
	if r.LibraryID == "" {
		return errors.New("library_id is required")
	}
	if r.Title == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}

type UpdateBookRequest struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"`
	CoverImage  string `json:"cover_image"`
}

func (r UpdateBookRequest) Valid() error {
	if r.ID == "" {
		return errors.New("id cannot be empty")
	}
	if r.Title == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}
