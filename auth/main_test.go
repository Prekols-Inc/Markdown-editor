package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func setupRouter() *gin.Engine {
	JWT_SECRET = []byte("test-secret")

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.POST("/v1/register", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{RSP_ERROR_KEY: "invalid request body"})
			return
		}

		if strings.TrimSpace(req.Username) == "" || strings.TrimSpace(req.Password) == "" {
			c.JSON(http.StatusBadRequest, gin.H{RSP_ERROR_KEY: "username and password are required"})
			return
		}

		token, err := generateToken(uuid.New())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{RSP_ERROR_KEY: "failed to generate jwt token"})
			return
		}

		c.SetCookie(TOKEN_COOKIE_NAME, token, 24*60*60, "/", "", false, true)
		c.JSON(http.StatusCreated, gin.H{RSP_MSG_KEY: "registration successful"})
	})

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

		adminUUID, err := uuid.Parse(UUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{RSP_ERROR_KEY: "internal server error"})
			return
		}

		token, err := generateToken(adminUUID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{RSP_ERROR_KEY: "failed to generate jwt token"})
			return
		}

		c.SetCookie(TOKEN_COOKIE_NAME, token, 24*60*60, "/", "", false, true)
		c.JSON(http.StatusOK, gin.H{RSP_MSG_KEY: "login successfull"})
	})

	return r
}

func TestRegisterHandler(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  any
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "Valid register",
			requestBody:  LoginRequest{Username: "new-user", Password: "new-password"},
			expectedCode: http.StatusCreated,
			expectedMsg:  "registration successful",
		},
		{
			name:         "Missing username",
			requestBody:  LoginRequest{Username: "", Password: "new-password"},
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "username and password are required",
		},
		{
			name:         "Invalid request body",
			requestBody:  42,
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid request body",
		},
	}

	r := setupRouter()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody, err := json.Marshal(test.requestBody)
			assert.Nil(t, err)

			req, _ := http.NewRequest(http.MethodPost, "/v1/register", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedCode, w.Code, "Check status code")
			assert.Contains(t, w.Body.String(), test.expectedMsg, "Check response body")
		})
	}
}

func TestLoginHandler(t *testing.T) {
	tests := []struct {
		name         string
		requestBody  any
		expectedCode int
		expectedMsg  string
	}{
		{
			name:         "Valid login",
			requestBody:  LoginRequest{Username: USERNAME, Password: PASSWORD},
			expectedCode: http.StatusOK,
			expectedMsg:  "login successfull",
		},
		{
			name:         "Invalid login",
			requestBody:  LoginRequest{Username: "invalid login", Password: PASSWORD},
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "invalid username or password",
		},
		{
			name:         "Invalid password",
			requestBody:  LoginRequest{Username: USERNAME, Password: "invalid password"},
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "invalid username or password",
		},
		{
			name:         "Invalid login and password",
			requestBody:  LoginRequest{Username: "invalid login", Password: "invalid password"},
			expectedCode: http.StatusUnauthorized,
			expectedMsg:  "invalid username or password",
		},
		{
			name:         "Invalid request body",
			requestBody:  "Invalid request body",
			expectedCode: http.StatusBadRequest,
			expectedMsg:  "invalid request body",
		},
	}

	r := setupRouter()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reqBody, err := json.Marshal(test.requestBody)
			assert.Nil(t, err)

			req, _ := http.NewRequest(http.MethodPost, "/v1/login", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, test.expectedCode, w.Code, "Check status code")
			assert.Contains(t, w.Body.String(), test.expectedMsg, "Check response body")
		})
	}
}
