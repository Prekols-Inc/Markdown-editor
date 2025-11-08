package repodb

import (
	"fmt"
	"strings"
	"testing"
)

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{"Valid filename", "test.md", true},
		{"Empty filename", "", false},
		{"Too long", string(make([]byte, 256)), false},
		{"Contains null", "file\x00name.md", false},
		{"Single dot", ".", false},
		{"Double dot", "..", false},
		{"Valid hidden file", ".gitignore.md", true},
		{"With spaces", "my file.md", true},
		{"With special chars", "file-name_v1.md", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidFilename(tt.filename); got != tt.want {
				t.Errorf("isValidLinuxFilename(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestValidateFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantErr  bool
		errMsg   string
	}{
		{"Valid md file", "test.md", false, ""},
		{"Valid txt file", "test.markdown", false, ""},
		{"Invalid extension", "test.jpg", true, fmt.Sprintf("Invalid filename: %s", ERR_BAD_EXTENSION)},
		{"Path traversal", "../test.md", true, fmt.Sprintf("Invalid filename: %s", ERR_PATH_IN_FILENAME)},
		{"Invalid filename", "file/name.md", true, fmt.Sprintf("Invalid filename: %s", ERR_PATH_IN_FILENAME)},
		{"Empty filename", "", true, fmt.Sprintf("Invalid filename: %s", ERR_INVALID_CHARACTERS)},
		{"Comma in filename", "file,md.md", true, fmt.Sprintf("Invalid filename: %s", ERR_INVALID_CHARACTERS)},
		{"Dot file with valid ext", ".env.md", false, ""},
		{"Only extension", ".md", true, fmt.Sprintf("Invalid filename: %s", ERR_EMPTY_FILENAME)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateFile(tt.filename)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateFile(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
				return
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if err.Error() != tt.errMsg && !strings.HasPrefix(err.Error(), tt.errMsg) {
					t.Errorf("validateFile(%q) error message = %v, want containing %v",
						tt.filename, err.Error(), tt.errMsg)
				}
			}
		})
	}
}
