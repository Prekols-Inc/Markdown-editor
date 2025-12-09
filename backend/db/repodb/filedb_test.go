package repodb

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testUUID uuid.UUID = uuid.New()

func setupTestDir(t *testing.T) (string, func()) {
	tempDir, err := os.MkdirTemp("", "filerepo_test")
	require.NoError(t, err)

	cleanup := func() { _ = os.RemoveAll(tempDir) }

	return tempDir, cleanup
}

func TestNewLocalFileRepo(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.DirExists(t, tempDir)

	nonExistentDir := filepath.Join(tempDir, "subdir")
	repo2, err := NewLocalFileRepo(nonExistentDir)
	assert.NoError(t, err)
	assert.NotNil(t, repo2)
	assert.DirExists(t, nonExistentDir)
}

func TestLocalFileRepo_CreateAndGet(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	filename := "test.md"
	content := []byte("Hello, World!")

	err = repo.Create(filename, testUUID, content)
	assert.NoError(t, err)

	filePath := filepath.Join(tempDir, testUUID.String(), filename)
	assert.FileExists(t, filePath)

	retrievedContent, err := repo.Get(filename, testUUID)
	assert.NoError(t, err)
	assert.Equal(t, content, retrievedContent)
}

func TestLocalFileRepo_CreateDuplicate(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	filename := "test.md"
	content := []byte("Hello, World!")

	err = repo.Create(filename, testUUID, content)
	assert.NoError(t, err)

	err = repo.Create(filename, testUUID, []byte("Different content"))
	assert.Error(t, err)
	assert.Equal(t, ErrFileExists, err)
}

func TestLocalFileRepo_GetNonExistent(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	_, err = repo.Get("nonexistent.md", testUUID)
	assert.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)
}

func TestLocalFileRepo_Save(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	filename := "test.md"
	initialContent := []byte("Initial content")
	updatedContent := []byte("Updated content")

	err = repo.Create(filename, testUUID, initialContent)
	assert.NoError(t, err)

	err = repo.Save(filename, testUUID, updatedContent)
	assert.NoError(t, err)

	retrievedContent, err := repo.Get(filename, testUUID)
	assert.NoError(t, err)
	assert.Equal(t, updatedContent, retrievedContent)
}

func TestLocalFileRepo_SaveNewFile(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	filename := "newfile.md"
	content := []byte("New content")

	err = repo.Save(filename, testUUID, content)
	assert.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)
}

func TestLocalFileRepo_Delete(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	filename := "test.md"
	content := []byte("Hello, World!")

	err = repo.Create(filename, testUUID, content)
	assert.NoError(t, err)

	err = repo.Delete(filename, testUUID)
	assert.NoError(t, err)

	_, err = repo.Get(filename, testUUID)
	assert.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)
}

func TestLocalFileRepo_DeleteNonExistent(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	err = repo.Delete("nonexistent.md", testUUID)
	assert.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)
}

func TestLocalFileRepo_GetList(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	files := map[string][]byte{
		"file1.md": []byte("Content 1"),
		"file2.md": []byte("Content 2"),
		"file3.md": []byte("Content 3"),
	}

	for filename, content := range files {
		err = repo.Create(filename, testUUID, content)
		assert.NoError(t, err)
	}

	titles, err := repo.GetList(testUUID)
	assert.NoError(t, err)
	assert.Len(t, titles, 3)

	expectedTitles := []string{"file1.md", "file2.md", "file3.md"}
	for _, expected := range expectedTitles {
		assert.Contains(t, titles, expected)
	}
}

func TestLocalFileRepo_GetListEmpty(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	titles, err := repo.GetList(testUUID)
	assert.NoError(t, err)
	assert.Empty(t, titles)
}

func TestLocalFileRepo_FilenameValidation(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	testCases := []struct {
		name        string
		filename    string
		shouldError bool
	}{
		{"Valid .md file", "test.md", false},
		{"Valid .markdown file", "test.markdown", false},
		{"Invalid extension", "test.txt", true},
		{"No extension", "test", true},
		{"Path traversal", "../test.md", true},
		{"Empty filename", "", true},
		{"Nested path", "subdir/test.md", true},
		{"Valid with dots", "test.v2.md", false},
		{"Valid with numbers", "123.md", false},
		{"Valid with underscores", "test_file.md", false},
		{"Valid with hyphens", "test-file.md", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := repo.Create(tc.filename, testUUID, []byte("content"))

			if tc.shouldError {
				assert.Error(t, err, "Expected error for filename: %s", tc.filename)
			} else {
				assert.NoError(t, err, "Expected no error for filename: %s", tc.filename)

				err = repo.Delete(tc.filename, testUUID)
				assert.NoError(t, err, "Expected no error for file: %s", tc.filename)
			}
		})
	}
}

func TestLocalFileRepo_FilePersistence(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo1, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	filename := "persistent.md"
	content := []byte("Persistent content")

	err = repo1.Create(filename, testUUID, content)
	assert.NoError(t, err)

	repo2, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	retrievedContent, err := repo2.Get(filename, testUUID)
	assert.NoError(t, err)
	assert.Equal(t, content, retrievedContent)
}

func TestLocalFileRepo_SpecialCharacters(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	specialFilenames := []string{
		"file with spaces.md",
		"file-with-dashes.md",
		"file_with_underscores.md",
		"file.with.dots.md",
		"UPPERCASE.MD",
		"MixedCase.Md",
	}

	for _, filename := range specialFilenames {
		t.Run(filename, func(t *testing.T) {
			content := []byte("content for " + filename)

			err := repo.Create(filename, testUUID, content)
			assert.NoError(t, err)

			retrieved, err := repo.Get(filename, testUUID)
			assert.NoError(t, err)
			assert.Equal(t, content, retrieved)

			err = repo.Delete(filename, testUUID)
			assert.NoError(t, err)
		})
	}
}

func TestFileNumberLimit(t *testing.T) {
	tempDir, cleanup := setupTestDir(t)
	defer cleanup()

	repo, err := NewLocalFileRepo(tempDir)
	require.NoError(t, err)

	for i := range MAX_USER_FILES {
		filename := fmt.Sprintf("file%d.md", i)
		t.Run(filename, func(t *testing.T) {
			content := []byte("content for " + filename)

			err := repo.Create(filename, testUUID, content)
			assert.NoError(t, err)

			retrieved, err := repo.Get(filename, testUUID)
			assert.NoError(t, err)
			assert.Equal(t, content, retrieved)
		})
	}

	content := []byte("content for last file")
	filename := "file_last.md"

	err = repo.Create(filename, testUUID, content)
	assert.Error(t, err)
	assert.Equal(t, ErrFileNumberLimitReached, err)

	_, err = repo.Get(filename, testUUID)
	assert.Error(t, err)
	assert.Equal(t, ErrFileNotFound, err)
}

func TestStorageLimit(t *testing.T) {
	for cnt := 1; cnt <= MAX_USER_FILES; cnt += 1 {
		t.Run(fmt.Sprintf("file count: %d", cnt), func(t *testing.T) {
			tempDir, cleanup := setupTestDir(t)
			defer cleanup()

			repo, err := NewLocalFileRepo(tempDir)
			require.NoError(t, err)

			size := USER_SPACE_SIZE/cnt + cnt
			content := make([]byte, size)
			for i := range content {
				content[i] = 1
			}

			for i := range cnt - 1 {
				filename := fmt.Sprintf("file%d.md", i)
				err = repo.Create(filename, testUUID, content[:size-cnt])
				assert.NoError(t, err)

				retrieved, err := repo.Get(filename, testUUID)
				assert.NoError(t, err)
				assert.Equal(t, content[:size-cnt], retrieved)
			}

			filename := "last_file.md"

			err = repo.Create(filename, testUUID, content)
			assert.Error(t, err)
			assert.Equal(t, ErrUserSpaceIsFull, err)

			_, err = repo.Get(filename, testUUID)
			assert.Error(t, err)
			assert.Equal(t, ErrFileNotFound, err)
		})
	}
}
