package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
<<<<<<< Updated upstream
	"os"
	"strconv"
=======
>>>>>>> Stashed changes
	"time"

	"backend/db/repodb"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

const (
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
	err := godotenv.Load("../.env")
	if err != nil {
		panic(fmt.Sprintf("Error loading .env file: %v", err))
	}
	port := os.Getenv("BACKEND_PORT")
	host := os.Getenv("BACKEND_HOST")

	if err := validatePort(port); err != nil {
		panic(fmt.Sprintf("Invalid port: %v\n", err))
	}

	router.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	repo, err := repodb.NewLocalFileRepo(DB_PATH)
	if err != nil {
		panic(fmt.Sprintf("Failed to create file repository: %v", err))
	}

	router := gin.Default()
	router.GET("/health", healthHandler)
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

	serverAddr := fmt.Sprintf("%s:%s", host, port)
	if err := router.Run(serverAddr); err != nil {
		panic(fmt.Sprintf("Failed to run server: %v", err))
	}
	fmt.Printf("Server started on %s\n", serverAddr)
}

func validatePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}

	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}

	if port <= 1023 {
		return fmt.Errorf("port %d is a system port and requires root privileges", port)
	}

	return nil
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}

func getFile(c *gin.Context) (*File, error) {
	file, _, err := c.Request.FormFile("file")
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

	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filename not provided"})
		return nil, err
	}

	return NewFile(filename, fileBytes), nil
}

func uploadFileHandler(c *gin.Context, repo repodb.FileRepository) {
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

func editFileHandler(c *gin.Context, repo repodb.FileRepository) {
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

func downloadFileHandler(c *gin.Context, repo repodb.FileRepository) {
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

func deleteFileHandler(c *gin.Context, repo repodb.FileRepository) {
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

func getAllFilesHandler(c *gin.Context, repo repodb.FileRepository) {
	fileNames, err := repo.GetList()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load file list: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"files": fileNames,
	})
}
