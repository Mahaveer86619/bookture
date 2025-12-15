package storage

import (
	"io"

	"github.com/Mahaveer86619/bookture/server/pkg/config"
)

type StorageService interface {
	Init() error
	SaveBookFile(bookID, volumeID string, file io.Reader) (string, error)
	GetPath(bookID, relativePath string) string
	HealthCheck() error
}

func NewStorageService() StorageService {
	cfg := config.AppConfig

	switch cfg.STORAGE_DRIVER {
	case "local":
		return &LocalStorage{
			basePath: cfg.STORAGE_PATH,
		}
	// Case "s3": return &S3Storage{...}
	default:
		return &LocalStorage{
			basePath: "./uploads",
		}
	}
}
