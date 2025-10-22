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
	c.SetCookie(ACCESS_TOKEN_COOKIE_NAME, accessToken, int(ACCESS_TOKEN_TTL.Seconds()), "/", "", false, true)
	c.SetCookie(REFRESH_TOKEN_COOKIE_NAME, refreshToken, int(REFRESH_TOKEN_TTL.Seconds()), "/auth/refresh", "", false, true)
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
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
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid username or password"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid username or password"})
		return
	}

	accessToken, refreshToken, err := generateTokens(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	setCookieTokens(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, LoginResponse{Message: "login successful", Token: accessToken})
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
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}

	var exists bool
	err := a.DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE username=$1)", req.Username).Scan(&exists)
	if err != nil {
		log.Println(err.Error())
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal server error (DB)"})
		return
	}
	if exists {
		c.JSON(http.StatusConflict, ErrorResponse{Error: "user already exists"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to hash password"})
		return
	}

	_, err = a.DB.Exec(context.Background(),
		"INSERT INTO users (username, password_hash, created_at) VALUES ($1, $2, $3)",
		req.Username, string(hashed), time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create user"})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{Message: "user registered successfully"})
}

// @Summary Check auth
// @Tags auth
// @Description Check if user authenticated
// @Produce json
// @Success 200 {object} CheckAuthResponse "Login response"
// @Failure 401 {object} ErrorResponse "Error response"
// @Router /v1/check_auth [get]
func (a *App) checkAuthHandler(c *gin.Context) {
	accessTokenStr, err := c.Cookie(ACCESS_TOKEN_COOKIE_NAME)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing access token"})
		return
	}

	_, err = parseToken(accessTokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired access token"})
		return
	}

	c.JSON(http.StatusOK, CheckAuthResponse{Authenticated: true})
}

func (a *App) refreshHandler(c *gin.Context) {
	refreshTokenStr, err := c.Cookie(REFRESH_TOKEN_COOKIE_NAME)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "missing refresh token"})
		return
	}

	claims, err := parseToken(refreshTokenStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid or expired refresh token"})
		return
	}

	userIdStr, exists := claims["user_id"]
	userId, ok := userIdStr.(uuid.UUID)
	if !exists || !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{Error: "Invalid token claims"})
		return
	}

	accessToken, refreshToken, err := generateTokens(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to generate token"})
		return
	}

	setCookieTokens(c, accessToken, refreshToken)
	c.JSON(http.StatusOK, LoginResponse{Message: "login successful", Token: accessToken})
}
