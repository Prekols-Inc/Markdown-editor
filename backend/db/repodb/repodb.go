package repodb

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"
)

var ErrFileNotFound = errors.New("file not found")
var ErrFileExists = errors.New("file already exists")
var validExtensions = []string{".md", ".markdown"}

type FileRepository interface {
	Save(filename string, data []byte) error
	Create(filename string, data []byte) error
	Get(filename string) ([]byte, error)
	Delete(filename string) error
	GetList() ([]string, error)
}

func isValidLinuxFilename(filename string) bool {
	if filename == "" || len(filename) > 255 {
		return false
	}

	if !utf8.ValidString(filename) {
		return false
	}

	if strings.Contains(filename, "\x00") {
		return false
	}

	if filename == "." || filename == ".." {
		return false
	}

	return true
}

func validateFile(filename string) error {
	if !isValidLinuxFilename(filename) {
		return fmt.Errorf("filename is not valid")
	}

	if filepath.Base(filename) != filename {
		return fmt.Errorf("filename must not contain file path")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(validExtensions, ext) {
		return fmt.Errorf("file extension must be .md")
	}

	if strings.TrimSuffix(filename, ext) == "" {
		return fmt.Errorf("file name must not be empty")
	}

	return nil
}
