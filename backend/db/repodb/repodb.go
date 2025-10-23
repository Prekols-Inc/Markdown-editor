package repodb

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	USER_SPACE_SIZE = 1 << 9
	MAX_USER_FILES  = 3
)

var ErrFileNotFound = errors.New("file not found")
var ErrFileExists = errors.New("file already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrUserSpaceIsFull = errors.New("user space is full")
var ErrFileNumberLimitReached = errors.New("file number limit has been reached")

type ErrInvalidFilename struct {
	Reason string
}

func (e *ErrInvalidFilename) Error() string {
	return fmt.Sprintf("Invalid filename: %s", e.Reason)
}

var validExtensions = []string{".md", ".markdown"}
var invalidFilenameChars = []string{"\x00", ":", "*", "?", "\"", "<", ">", "|", "+", ",", "!", "%", "@"}

type FileRepository interface {
	Save(filename string, userId uuid.UUID, data []byte) error
	Create(filename string, userId uuid.UUID, data []byte) error
	Get(filename string, userId uuid.UUID) ([]byte, error)
	Delete(filename string, userId uuid.UUID) error
	GetList(userId uuid.UUID) ([]string, error)
}

func containsChars(filename string, сhars []string) bool {
	for _, ch := range сhars {
		if strings.Contains(filename, ch) {
			return true
		}
	}

	return false
}

func isValidFilename(filename string) bool {
	if filename == "" || len(filename) > 255 {
		return false
	}

	if !utf8.ValidString(filename) {
		return false
	}

	if containsChars(filename, invalidFilenameChars) {
		return false
	}

	if filename == "." || filename == ".." {
		return false
	}

	return true
}

func validateFile(filename string) error {
	if !isValidFilename(filename) {
		return &ErrInvalidFilename{Reason: "invalid characters"}
	}

	if filepath.Base(filename) != filename {
		return &ErrInvalidFilename{Reason: "filename must not contain file path"}
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(validExtensions, ext) {
		return &ErrInvalidFilename{Reason: "file extension must be .md"}
	}

	if strings.TrimSuffix(filename, ext) == "" {
		return &ErrInvalidFilename{Reason: "file name must not be empty"}
	}

	return nil
}
