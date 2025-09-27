package repodb

import (
	"errors"
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
