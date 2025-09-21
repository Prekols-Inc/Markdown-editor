package main

import (
	"flag"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	USERNAME    = "admin"
	PASSWORD    = "password"
	RSP_MSG_KEY = "message"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
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
			c.JSON(http.StatusBadRequest, gin.H{RSP_MSG_KEY: "invalid request body"})
			return
		}

		if req.Username != USERNAME || req.Password != PASSWORD {
			c.JSON(http.StatusUnauthorized, gin.H{RSP_MSG_KEY: "invalid username or password"})
			return
		}

		c.JSON(http.StatusOK, gin.H{RSP_MSG_KEY: "login successful"})
	})
	r.Run(fmt.Sprintf("%s:%s", host, port))
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"time":   time.Now(),
	})
}
