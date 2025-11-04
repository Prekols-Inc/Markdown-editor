package main

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"time"

	"backend/db/repodb"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

const (
	DB_PATH       = "storage"
	TLS_CERT_FILE = "tls/cert.pem"
	TLS_KEY_FILE  = "tls/key.pem"
)

var (
	ErrUserIdNotFound    = errors.New("user_id not found in claims")
	ErrInvalidUserIdType = errors.New("invalid user_id type")
)

type File struct {
	name  string
	bytes []byte
}

func NewFile(name string, bytes []byte) *File {
	return &File{name: name, bytes: bytes}
}

func validatePort(portStr string) error {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}

	if port < 1024 || port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535")
	}

	return nil
}

// @title           Markdown backend
// @version         1.0
// @description     Backend for Markdown-editor

// @host            localhost:1234
// @BasePath        /
func main() {
	var host, port string
	flag.StringVar(&host, "host", "", "Host to bind")
	flag.StringVar(&port, "port", "", "Port to bind")
	flag.Parse()

	if err := validatePort(port); err != nil {
		panic(fmt.Sprintf("Invalid port: %v\n", err))
	}

	if host == "" {
		panic("Host not provided")
	}

	repo, err := repodb.NewLocalFileRepo(DB_PATH)
	if err != nil {
		panic(fmt.Sprintf("Failed to create file repository: %v", err))
	}

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.Static("/docs", "./docs")
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/docs/swagger.json")))
	r.GET("/health", healthHandler)

	authorized := r.Group("/api")
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

	serverAddr := fmt.Sprintf("%s:%s", host, port)
	if err := r.RunTLS(serverAddr, TLS_CERT_FILE, TLS_KEY_FILE); err != nil {
		panic(fmt.Sprintf("Failed to run server: %v", err))
	}
	fmt.Printf("Server started on %s\n", serverAddr)
}
