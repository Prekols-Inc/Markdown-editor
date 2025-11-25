package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"backend/db/repodb"
	"backend/db/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testToken string
var testUUID uuid.UUID

func setupTestRouter(repo repodb.FileRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	authorized := router.Group("/api")
	authorized.Use(authMiddleware())
	authorized.GET("/files", func(c *gin.Context) {
		getAllFilesHandler(c, repo)
	})
	authorized.GET("/file/:filename", func(c *gin.Context) {
		downloadFileHandler(c, repo)
	})
	authorized.POST("/file/:filename", func(c *gin.Context) {
		uploadFileHandler(c, repo)
	})
	authorized.PUT("/file/:filename", func(c *gin.Context) {
		editFileHandler(c, repo)
	})
	authorized.DELETE("/file/:filename", func(c *gin.Context) {
		deleteFileHandler(c, repo)
	})

	return router
}

type CleanupFunc func()

func generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET environment variable not set")
	}
	return token.SignedString([]byte(secret))
}

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

func LoadFile(t *testing.T, r *gin.Engine, repo repodb.FileRepository, testFilename string, testContent string) *httptest.ResponseRecorder {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(testContent))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("POST", "/api/file/"+testFilename, body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestMain(m *testing.M) {
	err := godotenv.Load()
	if err != nil {
		panic("Fail loading .env file")
	}

	testUUID = uuid.New()
	testToken, err = generateToken(testUUID)
	if err != nil {
		panic("Fail to generate test token")
	}

	code := m.Run()

	os.Exit(code)
}

func TestUploadFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	r := setupTestRouter(repo)

	testContent := "Hello, World!"
	testFilename := "test.md"

	w := LoadFile(t, r, repo, testFilename, testContent)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File uploaded successfully", response["message"])
	assert.Equal(t, testFilename, response["filename"])

	savedContent, err := repo.Get(testFilename, testUUID)
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

	err = repo.Create(testFilename, testUUID, []byte(testContent))
	assert.NoError(t, err)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", testFilename)
	assert.NoError(t, err)
	_, err = part.Write([]byte(testContentSaved))
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	req, err := http.NewRequest("PUT", "/api/file/"+testFilename, body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File saved successfully", response["message"])
	assert.Equal(t, testFilename, response["filename"])

	savedContent, err := repo.Get(testFilename, testUUID)
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

	req, err := http.NewRequest("PUT", "/api/file/"+noExistsFile, body)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	errObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "error field should be an object")

	msg, ok := errObj["message"].(string)
	require.True(t, ok, "error.message should be a string")

	assert.Contains(t, msg, "Файл не найден")
}

func TestDownloadFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	testContent := "Hello, World!"
	testFilename := "test.md"
	err = repo.Create(testFilename, testUUID, []byte(testContent))
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("GET", "/api/file/"+testFilename, nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

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

	req, err := http.NewRequest("GET", "/api/file/"+testFilename, nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	errObj, ok := response["error"].(map[string]interface{})
	require.True(t, ok, "error field should be an object")

	msg, ok := errObj["message"].(string)
	require.True(t, ok, "error.message should be a string")

	assert.Contains(t, msg, "Файл не найден")
}

func TestDeleteFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	testContent := "Hello, World!"
	testFilename := "test.md"
	err = repo.Create(testFilename, testUUID, []byte(testContent))
	assert.NoError(t, err)

	router := setupTestRouter(repo)

	req, err := http.NewRequest("DELETE", "/api/file/"+testFilename, nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "File deleted successfuly!", response["message"])
	assert.Equal(t, testFilename, response["filename"])

	_, err = repo.Get(testFilename, testUUID)
	assert.Error(t, err)
}

func TestUploadFileWithoutFile(t *testing.T) {
	repo, cleanup, err := getNewLocalFileTestRepo()
	assert.NoError(t, err)
	defer cleanup()

	router := setupTestRouter(repo)

	testFilename := "test.md"
	req, err := http.NewRequest("POST", "/api/file/"+testFilename, nil)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "multipart/form-data")

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

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
		err = repo.Create(filename, testUUID, []byte(content))
		assert.NoError(t, err)
	}

	router := setupTestRouter(repo)

	req, err := http.NewRequest("GET", "/api/files", nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

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

	req, err := http.NewRequest("GET", "/api/files", nil)
	assert.NoError(t, err)

	req.AddCookie(&http.Cookie{
		Name:  "access_token",
		Value: testToken,
		Path:  "/",
	})

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

func TestFileCreationLimit(t *testing.T) {
	r := utils.NewRateLimiter()
	testUUID := uuid.New()
	for range 5 {
		res := r.Allow(testUUID)
		assert.True(t, res, "action should be allowed for underlimit")
	}
	res := r.Allow(testUUID)
	assert.False(t, res, "action denied for overlimit")
}
