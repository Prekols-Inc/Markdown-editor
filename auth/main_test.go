package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"net/http/httptest"
	"testing"
)

func setupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"POST", "GET", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

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

		c.JSON(http.StatusOK, gin.H{RSP_ERROR_KEY: "login successful"})
	})

	return r
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
			expectedMsg:  "login successful",
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
			assert.Contains(t, w.Body.String(), test.expectedMsg, "Check request body")
		})
	}
}
