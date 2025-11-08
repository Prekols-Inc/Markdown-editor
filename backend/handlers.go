package main

import (
	"backend/db/repodb"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	requestsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
	)
)

func init() {
	prometheus.MustRegister(requestsTotal)
}

// @Summary Check server health
// @Tags health
// @Description Check if server respond
// @Produce json
// @Success 200 {object} HealthResponse "Server health status"
// @Router /health [get]
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
		Time:   time.Now(),
	})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "JWT not provided"})
			return
		}

		token, err := parseToken(tokenString)
		if err != nil || token == nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorExpired != 0 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Token has expired"})
					return
				}
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Wrong jwt"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userId, exists := claims["user_id"]
			if !exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
				return
			}

			c.Set("user_id", userId)

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
			return
		}
	}
}

func counterMiddleware() gin.HandlerFunc {
	return func(_ *gin.Context) {
		requestsTotal.Inc()
		fmt.Println("COUNT!")
	}
}

func getFile(c *gin.Context) *File {
	filename := c.Param("filename")
	if filename == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "Filename not provided"})
		return nil
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "File not provided"})
		return nil
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to read file: " + err.Error()})
		return nil
	}

	return NewFile(filename, fileBytes)
}

func getUserId(c *gin.Context) *uuid.UUID {
	userIdField, exists := c.Get("user_id")
	if !exists {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "user_id not found in jwt claims"})
		return nil
	}

	userIdStr, ok := userIdField.(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid type for user_id"})
		return nil
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "user_id parse error: " + err.Error()})
		return nil
	}

	return &userId
}

// @Summary Upload file
// @Tags files
// @Description Upload new file to server
// @Param filename path string true "Filename to save"
// @Param file formData file true "File to upload"
// @Produce json
// @Success 200 {object} UploadResponce "Upload responce"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
// @Failure 409 {object} ErrorResponce "Error responce"
// @Failure 500 {object} ErrorResponce "Error responce"
// @Router /api/file/{filename} [post]
func uploadFileHandler(c *gin.Context, repo repodb.FileRepository) {
	file := getFile(c)
	if file == nil {
		return
	}

	userId := getUserId(c)
	if userId == nil {
		return
	}
	fmt.Printf("filename: %s", file.name)
	if err := repo.Create(file.name, *userId, file.bytes); err != nil {
		if mapRepoErr(c, err, "name") {
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "Unexpected error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		Message:  "File uploaded successfully",
		Filename: file.name,
	})
}

// @Summary Edit file
// @Tags files
// @Description Send edited file to server
// @Param filename path string true "Filename to save"
// @Param file formData file true "File to save"
// @Produce json
// @Success 200 {object} EditResponce "Edit responce"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
// @Failure 404 {object} ErrorResponce "Error responce"
// @Failure 409 {object} ErrorResponce "Error responce"
// @Failure 500 {object} ErrorResponce "Error responce"
// @Router /api/file/{filename} [put]
func editFileHandler(c *gin.Context, repo repodb.FileRepository) {
	file := getFile(c)
	if file == nil {
		return
	}

	userId := getUserId(c)
	if userId == nil {
		return
	}

	if err := repo.Save(file.name, *userId, file.bytes); err != nil {
		if mapRepoErr(c, err, "name") {
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "Unexpected error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, EditResponse{
		Message:  "File saved successfully",
		Filename: file.name,
	})
}

// @Summary Download file
// @Tags files
// @Description Download a file by filename
// @Param filename path string true "Filename to download"
// @Produce octet-stream
// @Success 200 {file} file "File content"
// @Failure 400 {object} ErrorResponse "Error response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Failure 404 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /api/file/{filename} [get]
func downloadFileHandler(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	userId := getUserId(c)
	if userId == nil {
		return
	}

	bytes, err := repo.Get(filename, *userId)
	if err != nil {
		if mapRepoErr(c, err, "filename") {
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", bytes)
}

// @Summary Delete file
// @Tags files
// @Description Delete file from server
// @Produce json
// @Param filename path string true "Filename to delete"
// @Success 200 {object} DeleteResponse "Delete response"
// @Failure 400 {object} ErrorResponse "Error response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Failure 404 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /api/file/{filename} [delete]
func deleteFileHandler(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	userId := getUserId(c)
	if userId == nil {
		return
	}

	if err := repo.Delete(filename, *userId); err != nil {
		if mapRepoErr(c, err, "filename") {
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "Unexpected error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeleteResponse{
		Message:  "File deleted successfuly!",
		Filename: filename,
	})
}

// @Summary User files
// @Tags files
// @Description Get all user files from server
// @Produce json
// @Success 200 {object} ErrorResponse "Error response"
// @Failure 400 {object} ErrorResponse "Error response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /api/files [get]
func getAllFilesHandler(c *gin.Context, repo repodb.FileRepository) {
	userId := getUserId(c)
	if userId == nil {
		return
	}

	fileNames, err := repo.GetList(*userId)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to load file list: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetAllFilesResponse{
		Files: fileNames,
	})
}

// @Summary User files
// @Tags files
// @Description Get all user files from server
// @Produce json
// @Success 200 {object} ErrorResponse "Error response"
// @Failure 400 {object} ErrorResponse "Error response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /api/files [get]
func renameFileHandler(c *gin.Context, repo repodb.FileRepository) {
	oldFilename := c.Param("oldName")
	if oldFilename == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "Filename not provided"})
		return
	}
	newFilename := c.Param("newName")
	if newFilename == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, ErrorResponse{Error: "Filename not provided"})
		return
	}

	userId := getUserId(c)
	if userId == nil {
		return
	}

	err := repo.Rename(oldFilename, newFilename, *userId)
	if err != nil {
		if mapRepoErr(c, err, "newName") {
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{Error: "Unexpected error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, MessageReponse{
		Message: "File renamed successfully!",
	})
}

//

func abortRich(c *gin.Context, status int, code, msg, field string, details any) {
	c.AbortWithStatusJSON(status, ErrorResponse{
		Error: APIError{Code: code, Message: msg, Field: field, Details: details},
	})
}

func mapRepoErr(c *gin.Context, err error, field string) bool {
	if err == nil {
		return false
	}

	if errors.Is(err, repodb.ErrUserSpaceIsFull) {
		abortRich(c, http.StatusConflict, "USER_SPACE_FULL",
			"Недостаточно места в хранилище пользователя.", "", nil)
		return true
	}
	if errors.Is(err, repodb.ErrFileNumberLimitReached) {
		abortRich(c, http.StatusConflict, "FILE_COUNT_LIMIT",
			"Превышен лимит количества файлов.", "", map[string]int{"max": repodb.MAX_USER_FILES})
		return true
	}
	if errors.Is(err, repodb.ErrFileExists) {
		abortRich(c, http.StatusConflict, "FILE_ALREADY_EXISTS",
			"Файл с таким именем уже существует.", field, nil)
		return true
	}
	if errors.Is(err, repodb.ErrFileNotFound) {
		abortRich(c, http.StatusNotFound, "FILE_NOT_FOUND",
			"Файл не найден.", field, nil)
		return true
	}
	var inv *repodb.ErrInvalidFilename
	if errors.As(err, &inv) {
		code := "FILE_NAME_INVALID"
		msg := "Недопустимое имя файла."
		det := map[string]any{}

		switch inv.Reason {
		case repodb.ERR_EMPTY_FILENAME:
			code, msg = "FILE_NAME_EMPTY", "Имя файла не может быть пустым."
		case repodb.ERR_PATH_IN_FILENAME:
			code, msg = "FILE_NAME_PATH", "Имя файла не должно содержать путь или разделители каталогов."
		case repodb.ERR_BAD_EXTENSION:
			code, msg = "FILE_EXTENSION_INVALID", "Разрешены только расширения: .md ."
			det["allowedExtensions"] = []string{".md", ".markdown"}
		case repodb.ERR_TRAILING_DOT_SPACE:
			code, msg = "FILE_NAME_TRAILING", "Имя файла не должно заканчиваться точкой или пробелом."
		case repodb.ERR_TOO_LONG:
			code, msg = "FILE_NAME_TOO_LONG", "Слишком длинное имя файла."
			det["maxLen"] = 255
		case repodb.ERR_RESERVED:
			code, msg = "FILE_NAME_RESERVED", "Это имя зарезервировано системой."
			if inv.ReservedAs != "" {
				det["reserved"] = inv.ReservedAs
			}
		case repodb.ERR_ONLY_DOTS:
			code, msg = "FILE_NAME_ONLY_DOTS", "Имя файла не может состоять только из точек."
		case repodb.ERR_ONLY_SPACES:
			code, msg = "FILE_NAME_ONLY_SPACES", "Имя файла не может состоять только из пробелов."
		case repodb.ERR_INVALID_CHARACTERS:
			code, msg = "FILE_NAME_INVALID_CHARS", "Имя файла содержит недопустимые символы."
			if len(inv.InvalidRunes) > 0 {
				var chars []string
				for _, r := range inv.InvalidRunes {
					if r <= 0x1F || r == 0x7F {
						chars = append(chars, fmt.Sprintf("U+%02X", r))
					} else {
						chars = append(chars, string(r))
					}
				}
				det["invalid"] = chars
			}
		}

		abortRich(c, http.StatusBadRequest, code, msg, field, det)
		return true
	}

	return false
}
