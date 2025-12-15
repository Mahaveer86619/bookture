package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	basePath string
}

func (s *LocalStorage) Init() error {
	if _, err := os.Stat(s.basePath); os.IsNotExist(err) {
		if err := os.MkdirAll(s.basePath, 0755); err != nil {
			return fmt.Errorf("failed to create storage directory: %w", err)
		}
	}
	return nil
}

func (s *LocalStorage) SaveBookFile(bookID, volumeID string, file io.Reader) (string, error) {
	dirPath := filepath.Join(s.basePath, "book_"+bookID, "vol_"+volumeID)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create book directory: %w", err)
	}

	filePath := filepath.Join(dirPath, "source_file")

	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, file); err != nil {
		return "", fmt.Errorf("failed to write file content: %w", err)
	}

	return filePath, nil
}

func (s *LocalStorage) GetPath(bookID, relativePath string) string {
	return filepath.Join(s.basePath, "book_"+bookID, relativePath)
}

func (s *LocalStorage) HealthCheck() error {
	if _, err := os.Stat(s.basePath); os.IsNotExist(err) {
		return fmt.Errorf("storage path %s does not exist", s.basePath)
	}

	return nil
}
