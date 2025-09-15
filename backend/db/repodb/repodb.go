package repodb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrFileNotFound = errors.New("file not found")

type FileRepository interface {
	Save(filename string, data []byte) error
	Get(filename string) ([]byte, error)
	Delete(filename string) error
	GetList() ([]string, error)
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

	bytes, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrFileNotFound
		}
		return nil, err
	}

	return bytes, nil
}

func (l *LocalFileRepo) Delete(filename string) error {
	path := filepath.Join(l.basePath, filename)
	return os.Remove(path)
}

func (l *LocalFileRepo) GetList() ([]string, error) {
	files, err := os.ReadDir(l.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory %s: %w", l.basePath, err)
	}

	var fileNames []string
	for _, file := range files {
		if !file.IsDir() {
			fileNames = append(fileNames, file.Name())
		}
	}

	if fileNames == nil {
		fileNames = []string{}
	}

	return fileNames, nil
}
