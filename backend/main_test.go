package main

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"backend/db/repodb"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(repo repodb.FileRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/files", func(c *gin.Context) {
		getAllFilesHandler(c, repo)
	})
	router.GET("/file/:filename", func(c *gin.Context) {
		downloadFileHandler(c, repo)
	})
	router.POST("/file/:filename", func(c *gin.Context) {
		uploadFileHandler(c, repo)
	})
	router.PUT("/file/:filename", func(c *gin.Context) {
		editFileHandler(c, repo)
	})
	router.DELETE("/file/:filename", func(c *gin.Context) {
		deleteFileHandler(c, repo)
	})

	return router
}

type CleanupFunc func()

func getNewLocalFileTestRepo() (repodb.FileRepository, CleanupFunc, error) {
	tempDir, err := os.MkdirTemp("", "tmp")
	if err != nil {
		return nil, nil, err
	}

	repo, err := repodb.NewLocalFileRepo(tempDir)
	if err != nil {
		return nil, nil, err
	}

	return repo, func() { os.RemoveAll(tempDir) }, nil
}

func TestUploadFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	testContent := "Hello, World!"
	testFilename := "test.md"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(testContent))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/file/"+testFilename, body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File uploaded successfully", response["message"])
	assert.Equal(t, testFilename, response["filename"])

	savedContent, err := repo.Get(testFilename)
	assert.NoError(t, err)
	assert.Equal(t, testContent, string(savedContent))
}

func TestSaveFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	testContent := "Hello, World!"
	testContentSaved := "Goodbye, World!"
	testFilename := "test.md"

	err = repo.Create(testFilename, []byte(testContent))
	assert.NoError(t, err)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(testContentSaved))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "/file/"+testFilename, body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File saved successfully", response["message"])
	assert.Equal(t, testFilename, response["filename"])

	savedContent, err := repo.Get(testFilename)
	assert.NoError(t, err)
	assert.Equal(t, testContentSaved, string(savedContent))
}

func TestSaveFileNotFound(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	testContent := "Hello, World!"
	noExistsFile := "noexists.md"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", noExistsFile)
	assert.NoError(t, err)
	_, err = part.Write([]byte(testContent))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "/file/"+noExistsFile, body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, strings.Contains(response["error"].(string), "file not found"))
}

func TestDownloadFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	testContent := "Hello, World!"
	testFilename := "test.md"
	err = repo.Create(testFilename, []byte(testContent))
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("GET", "/file/"+testFilename, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename="+testFilename, w.Header().Get("Content-Disposition"))
	assert.Equal(t, testContent, w.Body.String())
}

func TestDownloadFileNotFound(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	testFilename := "nonexistent.md"

	req, err := http.NewRequest("GET", "/file/"+testFilename, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.True(t, strings.Contains(response["error"].(string), "File not found"))
}

func TestDeleteFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	testContent := "Hello, World!"
	testFilename := "test.md"
	err = repo.Create(testFilename, []byte(testContent))
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("DELETE", "/file/"+testFilename, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File deleted successfuly!", response["message"])
	assert.Equal(t, testFilename, response["filename"])

	_, err = repo.Get(testFilename)
	assert.Error(t, err)
}

func TestUploadFileWithoutFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	testFilename := "test.md"
	req, err := http.NewRequest("POST", "/file/"+testFilename, nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "multipart/form-data")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File not provided", response["error"])
}

func TestGetAllFiles(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	testFiles := map[string]string{
		"file1.md": "Content of file 1",
		"file2.md": "Content of file 2",
		"file3.md": "Content of file 3",
	}

	for filename, content := range testFiles {
		err = repo.Create(filename, []byte(content))
		assert.NoError(t, err)
	}

	router := setupTestRouter(repo)

	req, err := http.NewRequest("GET", "/files", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	files, exists := response["files"]
	assert.True(t, exists, "Response should contain 'files' field")

	filesSlice := files.([]interface{})
	var fileNames []string
	for _, file := range filesSlice {
		fileNames = append(fileNames, file.(string))
	}

	assert.Equal(t, 3, len(fileNames))

	expectedFiles := []string{"file1.md", "file2.md", "file3.md"}
	for _, expectedFile := range expectedFiles {
		assert.Contains(t, fileNames, expectedFile)
	}
}

func TestGetAllFilesEmpty(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	req, err := http.NewRequest("GET", "/files", nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	files, exists := response["files"]
	assert.True(t, exists, "Response should contain 'files' field")
	assert.Equal(t, 0, len(files.([]interface{})))
}
