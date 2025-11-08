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
	USER_SPACE_SIZE = 1 << 10 // 1 Кб
	MAX_USER_FILES  = 3
)

var ErrFileNotFound = errors.New("file not found")
var ErrFileExists = errors.New("file already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrUserSpaceIsFull = errors.New("user space is full")
var ErrFileNumberLimitReached = errors.New("file number limit has been reached")

const (
	ERR_INVALID_CHARACTERS = "filename contains invalid characters"
	ERR_PATH_IN_FILENAME   = "filename must not contain file path"
	ERR_BAD_EXTENSION      = "file extension must be .md"
	ERR_EMPTY_FILENAME     = "filename must not be empty"
	ERR_TRAILING_DOT_SPACE = "filename must not end with a dot or space"
	ERR_TOO_LONG           = "filename is too long"
	ERR_RESERVED           = "filename is reserved"
	ERR_ONLY_DOTS          = "filename must not consist of dots only"
	ERR_ONLY_SPACES        = "filename must not consist of spaces only"
)

type ErrInvalidFilename struct {
	Reason       string
	InvalidRunes []rune // какие недопустимые
	ReservedAs   string // если имя зарезервировано Windows
}

func (e *ErrInvalidFilename) Error() string {
	return fmt.Sprintf("Invalid filename: %s", e.Reason)
}

var validExtensions = []string{".md", ".markdown"}
var invalidRunes = []rune{'<', '>', ':', '"', '/', '\\', '|', '?', '*', '+', ',', '!', '%', '@'}
var reservedBaseNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

type FileRepository interface {
	Save(filename string, userId uuid.UUID, data []byte) error
	Create(filename string, userId uuid.UUID, data []byte) error
	Get(filename string, userId uuid.UUID) ([]byte, error)
	Delete(filename string, userId uuid.UUID) error
	GetList(userId uuid.UUID) ([]string, error)
	Rename(filename string, newFilename string, userId uuid.UUID) error
}

func containsChars(filename string, сhars []string) bool {
	for _, ch := range сhars {
		if strings.Contains(filename, ch) {
			return true
		}
	}

	return false
}

func validateFile(filename string) error {
	if filename == "" {
		return &ErrInvalidFilename{Reason: ERR_INVALID_CHARACTERS}
	}
	if strings.TrimSpace(filename) == "" {
		return &ErrInvalidFilename{Reason: ERR_ONLY_SPACES}
	}
	if len(filename) > 255 {
		return &ErrInvalidFilename{Reason: ERR_TOO_LONG}
	}
	if !utf8.ValidString(filename) {
		return &ErrInvalidFilename{Reason: ERR_INVALID_CHARACTERS}
	}
	if filepath.Base(filename) != filename {
		return &ErrInvalidFilename{Reason: ERR_PATH_IN_FILENAME}
	}
	ext := strings.ToLower(filepath.Ext(filename))
	if !slices.Contains(validExtensions, ext) {
		return &ErrInvalidFilename{Reason: ERR_BAD_EXTENSION}
	}
	base := strings.TrimSuffix(filename, ext)
	if base == "" {
		return &ErrInvalidFilename{Reason: ERR_EMPTY_FILENAME}
	}

	onlyDots := true
	for _, r := range base {
		if r != '.' {
			onlyDots = false
			break
		}
	}
	if onlyDots {
		return &ErrInvalidFilename{Reason: ERR_ONLY_DOTS}
	}
	if strings.HasSuffix(filename, ".") || strings.HasSuffix(filename, " ") {
		return &ErrInvalidFilename{Reason: ERR_TRAILING_DOT_SPACE}
	}
	if strings.HasSuffix(filename, ".") || strings.HasSuffix(filename, " ") {
		return &ErrInvalidFilename{Reason: ERR_TRAILING_DOT_SPACE}
	}

	if strings.TrimSpace(base) == "" {
		return &ErrInvalidFilename{Reason: ERR_ONLY_SPACES}
	}
	for _, r := range reservedBaseNames {
		if strings.EqualFold(base, r) {
			return &ErrInvalidFilename{Reason: ERR_RESERVED, ReservedAs: r}
		}
	}
	seen := make(map[rune]struct{})
	var bad []rune
	for _, ch := range filename {
		if ch <= 0x1F || ch == 0x7F {
			if _, ok := seen[ch]; !ok {
				seen[ch] = struct{}{}
				bad = append(bad, ch)
			}
			continue
		}
		for _, inv := range invalidRunes {
			if ch == inv {
				if _, ok := seen[ch]; !ok {
					seen[ch] = struct{}{}
					bad = append(bad, ch)
				}
				break
			}
		}
	}
	if len(bad) > 0 {
		return &ErrInvalidFilename{Reason: ERR_INVALID_CHARACTERS, InvalidRunes: bad}
	}
	return nil
}

func isValidFilename(filename string) bool {
	return validateFile(filename) == nil
}
