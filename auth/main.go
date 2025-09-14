package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const (
	PORT        = ":8080"
	USERNAME    = "admin"
	PASSWORD    = "password"
	RSP_MSG_KEY = "message"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func main() {
	r := gin.Default()
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
	r.Run(PORT)
}
