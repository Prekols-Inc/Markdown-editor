package repodb

import (
	"fmt"
	"os"
	"path/filepath"
)

type FileRepository interface {
	Save(filename string, data []byte) error
	Get(filename string) ([]byte, error)
	Delete(filename string) error
}

type LocalFileRepo struct {
	basePath string
}

func NewLocalFileRepo(basePath string) (*LocalFileRepo, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	return &LocalFileRepo{basePath: basePath}, nil
}

func (l *LocalFileRepo) Save(filename string, data []byte) error {
	path := filepath.Join(l.basePath, filename)
	return os.WriteFile(path, data, 0644)
}
func (l *LocalFileRepo) Get(filename string) ([]byte, error) {
	path := filepath.Join(l.basePath, filename)
	return os.ReadFile(path)
}
func (l *LocalFileRepo) Delete(filename string) error {
	path := filepath.Join(l.basePath, filename)
	return os.Remove(path)
}
