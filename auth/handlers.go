package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	USERNAME          = "admin"
	PASSWORD          = "password"
	UUID              = "123e4567-e89b-12d3-a456-426614174000"
	TOKEN_COOKIE_NAME = "access_token"
)

// @Summary Check auth health
// @Tags health
// @Description Check if auth respond
// @Produce json
// @Success 200 {object} HealthResponse "Server health status"
// @Router /health [get]
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
		Time:   time.Now(),
	})
}

// @Summary Sign in
// @Tags auth
// @Accept json
// @Produce json
// @Param login body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Login response"
// @Failure 400 {object} ErrorResponse "Error response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /v1/login [post]
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Username != USERNAME || req.Password != PASSWORD {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid username or password"})
		return
	}

	adminUUID, err := uuid.Parse(UUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error"})
		return
	}

	token, err := generateToken(adminUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate jwt token"})
		return
	}
	c.SetCookie(TOKEN_COOKIE_NAME, token, 24*60*60, "/", "", false, true)

	c.JSON(http.StatusOK, LoginResponse{Message: "login successfull", Token: token})
}

// @Summary Check auth
// @Tags auth
// @Description Check if user authenticated
// @Produce json
// @Success 200 {object} CheckAuthResponse "Login responce"
// @Failure 401 {object} ErrorResponse "Error responce"
// @Router /v1/check_auth [get]
func checkAuthHandler(c *gin.Context) {
	tokenStr, err := c.Cookie(TOKEN_COOKIE_NAME)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing token"})
		return
	}

	_, err = parseToken(tokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired token"})
		return
	}

	c.JSON(http.StatusOK, CheckAuthResponse{Authenticated: true})
}
