package main

import (
	"errors"
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

type File struct {
	name  string
	bytes []byte
}

func NewFile(name string, bytes []byte) *File {
	return &File{name: name, bytes: bytes}
}

func main() {
	router := gin.Default()

	repo, err := repodb.NewLocalFileRepo(DB_PATH)
	if err != nil {
		panic(fmt.Sprintf("Failed to create file repository: %v", err))
	}

	router.GET("/files", func(c *gin.Context) {
		getAllFiles(c, repo)
	})
	router.GET("/file/:filename", func(c *gin.Context) {
		downloadFile(c, repo)
	})
	router.POST("/file/:filename", func(c *gin.Context) {
		uploadFile(c, repo)
	})
	router.PUT("/file/:filename", func(c *gin.Context) {
		editFile(c, repo)
	})
	router.DELETE("/file/:filename", func(c *gin.Context) {
		deleteFile(c, repo)
	})

	if err := router.Run(PORT); err != nil {
		panic(fmt.Sprintf("Failed to run server: %v", err))
	}
	fmt.Printf("Server started on %s\n", PORT)
}

func getFile(c *gin.Context) (*File, error) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File not provided"})
		return nil, err
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file: " + err.Error()})
		return nil, err
	}

	filename := c.PostForm("filename")
	if filename == "" {
		filename = header.Filename
	}

	return NewFile(filename, fileBytes), nil
}

func uploadFile(c *gin.Context, repo repodb.FileRepository) {
	file, err := getFile(c)
	if err != nil {
		return
	}

	if err := repo.Create(file.name, file.bytes); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File uploaded successfully",
		"filename": file.name,
	})
}

func editFile(c *gin.Context, repo repodb.FileRepository) {
	file, err := getFile(c)
	if err != nil {
		return
	}
	if err := repo.Save(file.name, file.bytes); err != nil {
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

func downloadFile(c *gin.Context, repo repodb.FileRepository) {
	filename := c.Param("filename")

	bytes, err := repo.Get(filename)
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

func getAllFiles(c *gin.Context, repo repodb.FileRepository) {
	fileNames, err := repo.GetList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load file list: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": fileNames,
	})
}
