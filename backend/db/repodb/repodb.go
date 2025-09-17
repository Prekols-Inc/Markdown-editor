package repodb

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrFileNotFound = errors.New("file not found")
var ErrFileExists = errors.New("file already exists")

type FileRepository interface {
	Save(filename string, data []byte) error
	Create(filename string, data []byte) error
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

func IsFileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err != nil {
		fmt.Println(err.Error())
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (l *LocalFileRepo) Save(filename string, data []byte) error {
	path := filepath.Join(l.basePath, filename)
	ex, err := IsFileExists(path)
	if err != nil {
		return err
	}
	if !ex {
		return ErrFileNotFound
	}
	return os.WriteFile(path, data, 0644)
}

func (l *LocalFileRepo) Create(filename string, data []byte) error {
	path := filepath.Join(l.basePath, filename)
	ex, err := IsFileExists(path)
	if err != nil {
		return err
	}
	if ex {
		return ErrFileExists
	}
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
