package views

import (
	"errors"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/utils"
)

// Books
type BookView struct {
	ID          string    `json:"id"`
	LibraryID   string    `json:"library_id"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	Description string    `json:"description"`
	CoverImage  string    `json:"cover_image,omitempty"`
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
	for i := range books {
		views[i] = ToBookView(&books[i])
	}
	return views
}

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
		return errors.New("id is required")
	}
	return nil
}

// Volumes
type VolumeView struct {
	ID        string     `json:"id"`
	BookID    string     `json:"book_id"`
	Title     string     `json:"title"`
	Index     int        `json:"index"` // Volume number (1, 2, 3...)
	Status    string     `json:"status"`
	FilePath  *string    `json:"file_path,omitempty"`
	Uploaded  bool       `json:"uploaded"`
	ParsedAt  *time.Time `json:"parsed_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func ToVolumeView(v *models.Volume) VolumeView {
	var filePath *string
	if v.FilePath != "" {
		fp := v.FilePath
		filePath = &fp
	}

	return VolumeView{
		ID:        utils.MaskID(v.ID),
		BookID:    utils.MaskID(v.BookID),
		Title:     v.Title,
		Index:     v.Index,
		Status:    v.Status,
		FilePath:  filePath,
		Uploaded:  v.Uploaded,
		ParsedAt:  v.ParsedAt,
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

func ToVolumeViews(volumes []models.Volume) []VolumeView {
	views := make([]VolumeView, len(volumes))
	for i := range volumes {
		views[i] = ToVolumeView(&volumes[i])
	}
	return views
}

type CreateVolumeRequest struct {
	BookID string `json:"book_id"`
	Title  string `json:"title"`
	Index  int    `json:"index"`
}

func (r CreateVolumeRequest) Valid() error {
	if r.BookID == "" {
		return errors.New("book_id is required")
	}
	if r.Title == "" {
		return errors.New("title is required")
	}
	if r.Index <= 0 {
		return errors.New("index must be greater than 0")
	}
	return nil
}

type UpdateVolumeRequest struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

func (r UpdateVolumeRequest) Valid() error {
	if r.ID == "" {
		return errors.New("id is required")
	}
	if r.Title == "" {
		return errors.New("title cannot be empty")
	}
	return nil
}
