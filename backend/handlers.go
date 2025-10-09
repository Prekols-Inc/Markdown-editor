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
)

// @Summary Check server health
// @Tags health
// @Description Check if server respond
// @Produce json
// @Success 200 {object} HealthResponce "Server health status"
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

func getFile(c *gin.Context) *File {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Filename not provided"})
		return nil
	}

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "File not provided"})
		return nil
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to read file: " + err.Error()})
		return nil
	}

	return NewFile(filename, fileBytes)
}

func getUserId(c *gin.Context) *uuid.UUID {
	userIdField, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "user_id not found in jwt claims"})
		return nil
	}

	userIdStr, ok := userIdField.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "invalid type for user_id"})
		return nil
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "user_id parse error: " + err.Error()})
		return nil
	}

	return &userId
}

// @Summary Upload file
// @Tags files
// @Description Upload new file to server
// @Security AuthApiKey
// @Param filename path string true "Filename to save"
// @Param file formData file true "File to upload"
// @Produce json
// @Success 200 {object} UploadResponce "Upload responce"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
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
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create file: " + err.Error()})
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
// @Security AuthApiKey
// @Param filename path string true "Filename to save"
// @Param file formData file true "File to save"
// @Produce json
// @Success 200 {object} EditResponce "Edit responce"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
// @Failure 404 {object} ErrorResponce "Error responce"
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
		if errors.Is(err, repodb.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to save file: " + err.Error()})
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
// @Security ApiKeyAuth
// @Param filename path string true "Filename to download"
// @Produce octet-stream
// @Success 200 {file} file "File content"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
// @Failure 404 {object} ErrorResponce "Error responce"
// @Failure 500 {object} ErrorResponce "Error responce"
// @Router /api/file/{filename} [get]
func downloadFileHandler(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	userId := getUserId(c)
	if userId == nil {
		return
	}

	bytes, err := repo.Get(filename, *userId)
	if err != nil {
		if errors.Is(err, repodb.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: "File not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", bytes)
}

// @Summary Delete file
// @Tags files
// @Description Delete file from server
// @Security AuthApiKey
// @Produce json
// @Param filename path string true "Filename to delete"
// @Success 200 {object} DeleteResponce "Delete responce"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
// @Failure 404 {object} ErrorResponce "Error responce"
// @Failure 500 {object} ErrorResponce "Error responce"
// @Router /api/file/{filename} [delete]
func deleteFileHandler(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	userId := getUserId(c)
	if userId == nil {
		return
	}

	if err := repo.Delete(filename, *userId); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Failed to delete file: " + err.Error()})
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
// @Security AuthApiKey
// @Produce json
// @Success 200 {object} ErrorResponce "Error responce"
// @Failure 400 {object} ErrorResponce "Error responce"
// @Failure 401 {object} ErrorResponce "Error responce"
// @Failure 500 {object} ErrorResponce "Error responce"
// @Router /api/files [get]
func getAllFilesHandler(c *gin.Context, repo repodb.FileRepository) {
	userId := getUserId(c)
	if userId == nil {
		return
	}

	fileNames, err := repo.GetList(*userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to load file list: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, GetAllFilesResponse{
		Files: fileNames,
	})
}
