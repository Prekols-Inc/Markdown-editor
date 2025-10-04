package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	USERNAME        = "admin"
	PASSWORD        = "password"
	RSP_TOKEN_KEY   = "token"
	RSP_ERROR_KEY   = "error"
	adminUUIDString = "123e4567-e89b-12d3-a456-426614174000"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func generateToken(userID uuid.UUID) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID.String(),
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}

func main() {
	var host, port string
	flag.StringVar(&host, "host", "", "Host to bind")
	flag.StringVar(&port, "port", "", "Port to bind")
	flag.Parse()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins:  true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", healthHandler)
	r.POST("/v1/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{RSP_ERROR_KEY: "invalid request body"})
			return
		}

		if req.Username != USERNAME || req.Password != PASSWORD {
			c.JSON(http.StatusUnauthorized, gin.H{RSP_ERROR_KEY: "invalid username or password"})
			return
		}

		adminUUID, err := uuid.Parse(adminUUIDString)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{RSP_ERROR_KEY: "internal server error"})
			return
		}

		token, err := generateToken(adminUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{RSP_ERROR_KEY: "failed to generate jwt token"})
			return
		}

		c.JSON(http.StatusOK, gin.H{RSP_TOKEN_KEY: token})
	})
	serverAddr := fmt.Sprintf("%s:%s", host, port)
	if err := r.Run(serverAddr); err != nil {
		panic(fmt.Sprintf("Failed to run server: %v", err))
	}
	fmt.Printf("Server started on %s\n", serverAddr)
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}
