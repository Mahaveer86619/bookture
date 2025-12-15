package views

import (
	"errors"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/utils"
)

type LibraryView struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	Books     []BookView `json:"books"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func ToLibraryView(l *models.Library) LibraryView {
	view := LibraryView{
		ID:        utils.MaskID(l.ID),
		UserID:    utils.MaskID(l.UserID),
		Name:      l.Name,
		CreatedAt: l.CreatedAt,
		UpdatedAt: l.UpdatedAt,
		Books:     []BookView{},
	}

	if len(l.Books) > 0 {
		view.Books = ToBookViews(l.Books)
	}

	return view
}

func ToLibraryViews(libraries []models.Library) []LibraryView {
	views := make([]LibraryView, len(libraries))
	for i, l := range libraries {
		views[i] = ToLibraryView(&l)
	}
	return views
}

type CreateLibraryRequest struct {
	Name string `json:"name"`
}

func (r CreateLibraryRequest) Valid() error {
	if r.Name == "" {
		return errors.New("library name cannot be empty")
	}
	return nil
}

type UpdateLibraryRequest struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (r UpdateLibraryRequest) Valid() error {
	if r.ID == "" {
		return errors.New("id cannot be empty")
	}
	if r.Name == "" {
		return errors.New("library name cannot be empty")
	}
	return nil
}
