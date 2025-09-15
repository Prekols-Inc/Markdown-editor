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

	router.GET("/download/:filename", func(c *gin.Context) {
		downloadFile(c, repo)
	})
	router.GET("/files", func(c *gin.Context) {
		getAllFiles(c, repo)
	})
	router.POST("/upload", func(c *gin.Context) {
		uploadFile(c, repo)
	})
	router.DELETE("/download/:filename", func(c *gin.Context) {
		deleteFile(c, repo)
	})

	return router
}

func TestUploadFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_upload")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	testContent := "Hello, World!"
	testFilename := "test.txt"

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(testContent))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/upload", body)
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

func TestDownloadFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_download")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

	testContent := "Hello, World!"
	testFilename := "test.txt"
	err = repo.Save(testFilename, []byte(testContent))
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("GET", "/download/"+testFilename, nil)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/octet-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "attachment; filename="+testFilename, w.Header().Get("Content-Disposition"))
	assert.Equal(t, testContent, w.Body.String())
}

func TestDownloadFileNotFound(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_download_not_found")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	testFilename := "nonexistent.txt"

	req, err := http.NewRequest("GET", "/download/"+testFilename, nil)
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
	tempDir, err := os.MkdirTemp("", "test_delete")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

	testContent := "Hello, World!"
	testFilename := "test.txt"
	err = repo.Save(testFilename, []byte(testContent))
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("DELETE", "/download/"+testFilename, nil)
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
	tempDir, err := os.MkdirTemp("", "test_no_file")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("POST", "/upload", nil)
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
	tempDir, err := os.MkdirTemp("", "test_get_all_files")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

	testFiles := map[string]string{
		"file1.txt": "Content of file 1",
		"file2.txt": "Content of file 2",
		"file3.txt": "Content of file 3",
	}

	for filename, content := range testFiles {
		err = repo.Save(filename, []byte(content))
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

	expectedFiles := []string{"file1.txt", "file2.txt", "file3.txt"}
	for _, expectedFile := range expectedFiles {
		assert.Contains(t, fileNames, expectedFile)
	}
}

func TestGetAllFilesEmpty(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test_get_all_files_empty")
	assert.NoError(t, err)
	defer os.RemoveAll(tempDir)

	repo, err := repodb.NewLocalFileRepo(tempDir)
	assert.NoError(t, err)

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
