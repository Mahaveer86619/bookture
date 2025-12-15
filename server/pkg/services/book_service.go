package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/enums"
	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/services/storage"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
	"gorm.io/gorm"
)

type BookService struct {
	db         *gorm.DB
	ss         storage.StorageService
	libService *LibraryService
	processor  *ProcessingService
	parser     *ParserService
}

func NewBookService(ss storage.StorageService, lib *LibraryService, proc *ProcessingService, parser *ParserService) *BookService {
	return &BookService{
		db:         db.GetBooktureDB().DB,
		ss:         ss,
		libService: lib,
		processor:  proc,
		parser:     parser,
	}
}

func (bs *BookService) CreateDraftBook(userID, libID uint, title string, author string, description string) (*views.BookView, error) {
	_, err := bs.libService.GetLibrary(libID, userID)
	if err != nil {
		return nil, err
	}

	book := models.Book{
		LibraryID:   libID,
		Title:       title,
		Author:      author,
		Description: description,
		Status:      enums.Draft.ToString(),
	}

	if err := bs.db.Create(&book).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to create book", err)
	}

	v := views.ToBookView(&book)
	return &v, nil
}

func (bs *BookService) UploadVolume(userID uint, bookID uint, fileHeader *multipart.FileHeader) (*views.VolumeView, error) {
	var book models.Book
	if err := bs.db.First(&book, bookID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "Book not found", err)
		}
		return nil, err
	}

	if _, err := bs.libService.GetLibrary(book.LibraryID, userID); err != nil {
		return nil, err
	}

	var count int64
	bs.db.Model(&models.Volume{}).Where("book_id = ?", bookID).Count(&count)
	newIndex := int(count) + 1

	inferredTitle := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))

	volume := models.Volume{
		BookID: bookID,
		Title:  inferredTitle,
		Index:  newIndex,
		Status: enums.Processing.ToString(),
	}

	if err := bs.db.Create(&volume).Error; err != nil {
		return nil, err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return nil, errz.New(errz.BadRequest, "Unable to open uploaded file", err)
	}
	defer file.Close()

	storagePath, err := bs.ss.SaveBookFile(fmt.Sprintf("%d", book.ID), fmt.Sprintf("%d", volume.ID), file)
	if err != nil {
		bs.db.Delete(&volume)
		return nil, errz.New(errz.InternalServerError, "Failed to save file to storage", err)
	}

	volume.FilePath = storagePath
	volume.Uploaded = true
	if err := bs.db.Save(&volume).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to update volume record", err)
	}

	jobID := fmt.Sprintf("vol-%d", volume.ID)

	bs.processor.Enqueue(jobID, func(reportProgress func(int)) {
		wrappedReporter := func(p int) {
			reportProgress(p)
			bs.db.Model(&models.Volume{}).Where("id = ?", volume.ID).Update("progress", p)
		}
		bs.parser.ParseVolumeMetadata(volume.ID, wrappedReporter)
	})

	v := views.ToVolumeView(&volume)
	return &v, nil
}

func (bs *BookService) GetBook(userID uint, bookID uint) (*views.BookView, error) {
	var book models.Book
	err := bs.db.Joins("JOIN libraries ON libraries.id = books.library_id").
		Where("books.id = ? AND libraries.user_id = ?", bookID, userID).
		First(&book).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "Book not found", err)
		}
		return nil, err
	}

	v := views.ToBookView(&book)
	return &v, nil
}
