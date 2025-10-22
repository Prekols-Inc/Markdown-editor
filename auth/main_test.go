package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockApp struct {
	loginFunc    func(req LoginRequest, c *gin.Context)
	registerFunc func(req RegisterRequest, c *gin.Context)
}

func (m *mockApp) loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request body"})
		return
	}
	m.loginFunc(req, c)
}

func (m *mockApp) registerHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	m.registerFunc(req, c)
}

func setupRouter(app *mockApp) *gin.Engine {
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	r.POST("/v1/login", app.loginHandler)
	r.POST("/v1/register", app.registerHandler)
	return r
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userID := uuid.MustParse(UUID)

	tests := []struct {
		name         string
		requestBody  any
		loginFunc    func(req LoginRequest, c *gin.Context)
		expectedCode int
		expectedMsg  string
	}{
		{
			name: "Valid login",
			requestBody: LoginRequest{
				Username: USERNAME,
				Password: PASSWORD,
			},
			loginFunc: func(req LoginRequest, c *gin.Context) {
				c.JSON(http.StatusOK, LoginResponse{
					Message:     "login successful",
					AccessToken: "fake_token_" + userID.String(),
				})
			},
			expectedCode: http.StatusOK,
			expectedMsg:  "login successful",
		},
		{
			name: "Invalid login",
			requestBody: LoginRequest{
				Username: "invalid",
				Password: PASSWORD,
			},
			loginFunc: func(req LoginRequest, c *gin.Context) {
				c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "invalid username or password"})
			},
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "invalid username or password",
		},
		{
			name:         "Invalid request body",
			requestBody:  "not-json",
			loginFunc:    func(req LoginRequest, c *gin.Context) {},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &mockApp{loginFunc: tt.loginFunc}
			r := setupRouter(app)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/v1/login", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, "status code mismatch")
			assert.Contains(t, w.Body.String(), tt.expectedMsg, "message mismatch")
		})
	}
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		requestBody  any
		registerFunc func(req RegisterRequest, c *gin.Context)
		expectedCode int
		expectedMsg  string
	}{
		{
			name: "Successful registration",
			requestBody: RegisterRequest{
				Username: "newuser",
				Password: "newpassword",
			},
			registerFunc: func(req RegisterRequest, c *gin.Context) {
				c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully"})
			},
			expectedCode: http.StatusCreated,
			expectedMsg:  "User registered successfully",
		},
		{
			name: "User already exists",
			requestBody: RegisterRequest{
				Username: "existing_user",
				Password: "password",
			},
			registerFunc: func(req RegisterRequest, c *gin.Context) {
				c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			},
			expectedCode: http.StatusConflict,
			expectedMsg:  "User already exists",
		},
		{
			name:         "Invalid request body",
			requestBody:  "invalid json",
			registerFunc: func(req RegisterRequest, c *gin.Context) {},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid request body",
		},
		{
			name: "Database error",
			requestBody: RegisterRequest{
				Username: "user",
				Password: "pass",
			},
			registerFunc: func(req RegisterRequest, c *gin.Context) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
			},
			expectedCode: http.StatusInternalServerError,
			expectedMsg:  "DB error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &mockApp{registerFunc: tt.registerFunc}
			r := setupRouter(app)

			reqBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest(http.MethodPost, "/v1/register", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedCode, w.Code, "status code mismatch")
			assert.Contains(t, w.Body.String(), tt.expectedMsg, "message mismatch")
		})
	}
}
