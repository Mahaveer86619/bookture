package services

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/Mahaveer86619/bookture/server/pkg/db"
	"github.com/Mahaveer86619/bookture/server/pkg/enums"
	"github.com/Mahaveer86619/bookture/server/pkg/errz"
	"github.com/Mahaveer86619/bookture/server/pkg/models"
	"github.com/Mahaveer86619/bookture/server/pkg/services/parser"
	"github.com/Mahaveer86619/bookture/server/pkg/services/storage"
	"github.com/Mahaveer86619/bookture/server/pkg/views"
	"gorm.io/gorm"
)

type BookService struct {
	db         *gorm.DB
	ss         storage.StorageService
	libService *LibraryService
	processor  *ProcessingService
	parser     *parser.ParserService
}

func NewBookService(
	ss storage.StorageService,
	lib *LibraryService,
	proc *ProcessingService,
	parser *parser.ParserService,
) *BookService {
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
		LibraryID:        libID,
		Title:            title,
		Author:           author,
		Description:      description,
		Status:           enums.BookDraft.ToString(),
		TotalVolumes:     0,
		CompletedVolumes: 0,
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
	fileFormat := strings.TrimPrefix(strings.ToLower(filepath.Ext(fileHeader.Filename)), ".")

	volume := models.Volume{
		BookID:     bookID,
		Title:      inferredTitle,
		Index:      newIndex,
		Status:     enums.VolumeCreated.ToString(),
		FileFormat: fileFormat,
		Progress:   0,
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

	now := time.Now()
	volume.FilePath = storagePath
	volume.Uploaded = true
	volume.UploadedAt = &now
	volume.Status = enums.VolumeUploaded.ToString()

	if err := bs.db.Save(&volume).Error; err != nil {
		return nil, errz.New(errz.InternalServerError, "Failed to update volume record", err)
	}

	// Update book status
	if book.Status == enums.BookDraft.ToString() {
		book.Status = enums.BookProcessing.ToString()
		book.TotalVolumes = newIndex
		bs.db.Save(&book)
	}

	jobID := fmt.Sprintf("parse-vol-%d", volume.ID)

	bs.processor.Enqueue(jobID, func(reportProgress func(int)) {
		wrappedReporter := func(p int) {
			reportProgress(p)
			bs.db.Model(&models.Volume{}).Where("id = ?", volume.ID).Update("progress", p)
		}
		bs.parser.ProcessVolumeComplete(volume.ID, wrappedReporter)
	})

	v := views.ToVolumeView(&volume)
	v.TaskID = jobID

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

func (bs *BookService) GetTaskProgress(taskID string) (int, error) {
	progress := bs.processor.GetProgress(taskID)
	return progress, nil
}

func (bs *BookService) GetVolumeDetails(userID uint, volumeID uint) (*views.VolumeDetailView, error) {
	var volume models.Volume

	// 1. Verify ownership (via Book -> Library -> User)
	// 2. Preload Chapters and Sections in order
	err := bs.db.
		Joins("JOIN books ON books.id = volumes.book_id").
		Joins("JOIN libraries ON libraries.id = books.library_id").
		Where("volumes.id = ? AND libraries.user_id = ?", volumeID, userID).
		Preload("Book").
		Preload("Chapters", func(db *gorm.DB) *gorm.DB {
			return db.Order("chapters.chapter_no ASC")
		}).
		Preload("Chapters.Sections", func(db *gorm.DB) *gorm.DB {
			return db.Order("sections.section_no ASC")
		}).
		Preload("Chapters.Sections.Scenes").
		First(&volume).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errz.New(errz.NotFound, "Volume not found", err)
		}
		return nil, err
	}

	// 3. Map to View
	view := &views.VolumeDetailView{
		ID:           volume.ID,
		Title:        volume.Title,
		Author:       volume.Book.Author,
		Status:       volume.Status,
		Progress:     volume.Progress,
		ChapterCount: len(volume.Chapters),
		WordCount:    volume.WordCount,
		Chapters:     make([]views.ChapterView, len(volume.Chapters)),
	}

	for i, ch := range volume.Chapters {
		chView := views.ChapterView{
			ID:        ch.ID,
			ChapterNo: ch.ChapterNo,
			Title:     ch.Title,
			WordCount: ch.WordCount,
			Sections:  make([]views.SectionView, len(ch.Sections)),
		}

		for j, sec := range ch.Sections {
			secView := views.SectionView{
				ID:          sec.ID,
				SectionNo:   sec.SectionNo,
				Content:     sec.CleanText,
				WordCount:   sec.WordCount,
				HasDialogue: sec.HasDialogue,
				HasAction:   sec.HasAction,
				Scenes:      make([]views.SceneView, len(sec.Scenes)),
			}

			for k, sc := range sec.Scenes {
				secView.Scenes[k] = views.SceneView{
					ID:              sc.ID,
					Summary:         sc.Summary,
					ImageURL:        sc.ImageURL,
					ImagePrompt:     sc.ImagePrompt,
					ImportanceScore: sc.ImportanceScore,
					SceneType:       sc.SceneType,
					Location:        sc.Location,
					Mood:            sc.Mood,
				}
			}
			chView.Sections[j] = secView
		}
		view.Chapters[i] = chView
	}

	return view, nil
}
