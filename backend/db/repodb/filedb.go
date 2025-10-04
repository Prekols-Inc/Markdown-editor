package repodb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
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

func getPath(basePath string, userId uuid.UUID, filename string) (string, error) {
	err := validateFile(filename)
	if err != nil {
		return "", err
	}
	path := filepath.Join(basePath, userId.String(), filename)

	return path, nil
}

func createUserDirIfNotExists(basePath string, userId uuid.UUID) error {
	path := filepath.Join(basePath, userId.String())
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", path, err)
			}

			return nil
		}

		return fmt.Errorf("failed to create user dir %s: %w", path, err)
	}

	return nil
}

func (l *LocalFileRepo) Save(filename string, userId uuid.UUID, data []byte) error {
	path, err := getPath(l.basePath, userId, filename)
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

func (l *LocalFileRepo) Create(filename string, userId uuid.UUID, data []byte) error {
	err := createUserDirIfNotExists(l.basePath, userId)
	if err != nil {
		return err
	}

	path, err := getPath(l.basePath, userId, filename)
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

func (l *LocalFileRepo) Get(filename string, userId uuid.UUID) ([]byte, error) {
	path, err := getPath(l.basePath, userId, filename)
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

func (l *LocalFileRepo) Delete(filename string, userId uuid.UUID) error {
	path, err := getPath(l.basePath, userId, filename)
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

func (l *LocalFileRepo) GetList(userId uuid.UUID) ([]string, error) {
	err := createUserDirIfNotExists(l.basePath, userId)
	if err != nil {
		return nil, err
	}

	path := filepath.Join(l.basePath, userId.String())
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory %s: %w", path, err)
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
