package repodb

import (
	"fmt"
	"os"
	"path/filepath"
)

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
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func getPath(basePath string, filename string) (string, error) {
	err := validateFile(filename)
	if err != nil {
		return "", err
	}
	path := filepath.Join(basePath, filename)

	return path, nil
}

func (l *LocalFileRepo) Save(filename string, data []byte) error {
	path, err := getPath(l.basePath, filename)
	if err != nil {
		return err
	}

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
	path, err := getPath(l.basePath, filename)
	if err != nil {
		return err
	}

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
	path, err := getPath(l.basePath, filename)
	if err != nil {
		return nil, err
	}

	ex, err := IsFileExists(path)
	if err != nil {
		return nil, err
	}
	if !ex {
		return nil, ErrFileNotFound
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func (l *LocalFileRepo) Delete(filename string) error {
	path, err := getPath(l.basePath, filename)
	if err != nil {
		return err
	}

	ex, err := IsFileExists(path)
	if err != nil {
		return err
	}
	if !ex {
		return ErrFileNotFound
	}

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
