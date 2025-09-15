package main

import (
	"fmt"
	"io"
	"net/http"

	"backend/db/repodb"

	"github.com/gin-gonic/gin"
)

const (
	PORT    = ":1234"
	DB_PATH = "db/storage"
)

func main() {
	router := gin.Default()

	repo, err := repodb.NewLocalFileRepo(DB_PATH)
	if err != nil {
		panic(fmt.Sprintf("Failed to create file repository: %v", err))
	}

	router.POST("/upload", func(c *gin.Context) {
		uploadFile(c, repo)
	})
	router.GET("/download/:filename", func(c *gin.Context) {
		downloadFile(c, repo)
	})
	router.DELETE("download/:filename", func(c *gin.Context) {
		deleteFile(c, repo)
	})

	if err := router.Run(PORT); err != nil {
		panic(fmt.Sprintf("Failed to run server: %v", err))
	}
	fmt.Printf("Server started on %s\n", PORT)
}

func uploadFile(c *gin.Context, repo repodb.FileRepository) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
		return
	}

	filename := c.PostForm("filename")
	if filename == "" {
		filename = header.Filename
	}

	if err := repo.Save(filename, fileBytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": filename,
	})
}

func downloadFile(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	bytes, err := repo.Get(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found: " + err.Error()})
		return
	}

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", bytes)
}

func deleteFile(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	if err := repo.Delete(filename); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to delete file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File deleted successfuly!",
		"filename": filename,
	})
}
