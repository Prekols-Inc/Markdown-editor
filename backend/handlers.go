package main

import (
	"backend/db/repodb"
	"errors"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}

func parseToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("JWT_SECRET")), nil
	})
}

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil || tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "JWT not provided"})
			return
		}

		token, err := parseToken(tokenString)
		if err != nil || token == nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				if ve.Errors&jwt.ValidationErrorExpired != 0 {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Token has expired"})
					return
				}
			}

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Wrong jwt"})
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			userId, exists := claims["user_id"]
			if !exists {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
				return
			}

			c.Set("user_id", userId)

			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			return
		}
	}
}

func getFile(c *gin.Context) *File {
	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return nil
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
		return nil
	}

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename not provided"})
		return nil
	}

	return NewFile(filename, fileBytes)
}

func getUserId(c *gin.Context) *uuid.UUID {
	userIdField, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id not found in jwt claims"})
		return nil
	}

	userIdStr, ok := userIdField.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid type for user_id"})
		return nil
	}

	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "user_id parse error: " + err.Error()})
		return nil
	}

	return &userId
}

func uploadFileHandler(c *gin.Context, repo repodb.FileRepository) {
	file := getFile(c)
	if file == nil {
		return
	}

	userId := getUserId(c)
	if userId == nil {
		return
	}

	if err := repo.Create(file.name, *userId, file.bytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": file.name,
	})
}

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
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File saved successfully",
		"filename": file.name,
	})
}

func downloadFileHandler(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	userId := getUserId(c)
	if userId == nil {
		return
	}

	bytes, err := repo.Get(filename, *userId)
	if err != nil {
		if errors.Is(err, repodb.ErrFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", bytes)
}

func deleteFileHandler(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	userId := getUserId(c)
	if userId == nil {
		return
	}

	if err := repo.Delete(filename, *userId); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File deleted successfuly!",
		"filename": filename,
	})
}

func getAllFilesHandler(c *gin.Context, repo repodb.FileRepository) {
	userId := getUserId(c)
	if userId == nil {
		return
	}

	fileNames, err := repo.GetList(*userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load file list: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": fileNames,
	})
}
