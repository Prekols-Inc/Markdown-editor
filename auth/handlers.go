package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	USERNAME                  = "admin"
	PASSWORD                  = "password"
	UUID                      = "123e4567-e89b-12d3-a456-426614174000"
	ACCESS_TOKEN_COOKIE_NAME  = "access_token"
	REFRESH_TOKEN_COOKIE_NAME = "refresh_token"
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

func setCookieTokens(c *gin.Context, accessToken string, refreshToken string) {
	c.SetCookie(ACCESS_TOKEN_COOKIE_NAME, accessToken, int(REFRESH_TOKEN_TTL.Seconds()), "/", "", false, true)
	c.SetCookie(REFRESH_TOKEN_COOKIE_NAME, refreshToken, int(REFRESH_TOKEN_TTL.Seconds()), "/v1/refresh", "", false, true)
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
func (a *App) loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	var (
		id           uuid.UUID
		passwordHash string
	)

	err := a.DB.QueryRow(context.Background(),
		"SELECT id, password_hash FROM users WHERE username=$1", req.Username).
		Scan(&id, &passwordHash)

	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid username or password"})
		return
	}

	accessToken, refreshToken, err := generateTokens(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	setCookieTokens(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, LoginResponse{Message: "Login successful"})
}

func (a *App) logoutHandler(c *gin.Context) {
	c.SetCookie(ACCESS_TOKEN_COOKIE_NAME, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, LogoutResponse{Message: "logout successful"})
}

// @Summary Check auth
// @Tags auth
// @Description Check if user authenticated
// @Produce json
// @Success 200 {object} CheckAuthResponse "Login response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Router /v1/check_auth [get]
func (a *App) checkAuthHandler(c *gin.Context) {
	tokenStr, err := c.Cookie(ACCESS_TOKEN_COOKIE_NAME)
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

// @Summary Register
// @Tags auth
// @Description Register new user
// @Accept json
// @Produce json
// @Param register body RegisterRequest true "Register fields"
// @Success 201 {object} RegisterResponse "Login response"
// @Failure 400 {object} ErrorResponse "Error response"
// @Failure 409 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /v1/register [post]
func (a *App) registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "Invalid request body"})
		return
	}

	var exists bool
	err := a.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", req.Username).Scan(&exists)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Internal server error (DB)"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "User already exists"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to hash password"})
		return
	}

	_, err = a.DB.Exec(context.Background(),
		"INSERT INTO users (username, password_hash, created_at) VALUES ($1, $2, $3)",
		req.Username, string(hashed), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{Message: "User registered successfully"})
}

// @Summary Refresh tokens
// @Tags auth
// @Description Refresh access and refresh tokens
// @Produce json
// @Success 200 {object} LoginResponse "Login response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Failure 500 {object} ErrorResponse "Error response"
// @Router /v1/refresh [post]
func (a *App) refreshHandler(c *gin.Context) {
	refreshTokenStr, err := c.Cookie(REFRESH_TOKEN_COOKIE_NAME)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Missing refresh token"})
		return
	}

	claims, err := parseToken(refreshTokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid or expired refresh token"})
		return
	}

	userIdObj, exists := claims["user_id"]
	if !exists {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
		return
	}
	userIdStr, ok := userIdObj.(string)
	if !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
		return
	}
	userId, err := uuid.Parse(userIdStr)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
		return
	}

	accessToken, refreshToken, err := generateTokens(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "Failed to generate token"})
		return
	}

	setCookieTokens(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, RefreshResponse{Message: "Refresh success"})
}
