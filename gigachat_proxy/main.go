package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	TLS_CERT_FILE = "cert.crt"
	TLS_KEY_FILE  = "key.crt"
)

var gigaToken string
var gigaTokenExpires time.Time

func main() {
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/health", healthHandler)
	r.POST("/api/gigachat/summarize", summarizeHandler)

	host := os.Getenv("GIGACHAT_PROXY_HOST")
	port := os.Getenv("GIGACHAT_PROXY_PORT")

	serverAddr := fmt.Sprintf("%s:%s", host, port)
	tlsDir := os.Getenv("CERT_DIR_PATH")
	if err := r.RunTLS(serverAddr, tlsDir+TLS_CERT_FILE, tlsDir+TLS_KEY_FILE); err != nil {
		panic(fmt.Sprintf("Failed to run proxy: %v", err))
	}
	fmt.Printf("Gigachat Proxy is running on %s\n", serverAddr)
}

func healthHandler(c *gin.Context) {
	type HealthResponse struct {
		Status string    `json:"status"`
		Time   time.Time `json:"time"`
	}

	c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
		Time:   time.Now(),
	})
}

func summarizeHandler(c *gin.Context) {
	var req struct {
		Text string `json:"text"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	token, err := getGigachatToken()
	if err != nil {
		log.Println("Error fetching token:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "token fetch failed"})
		return
	}

	summary, err := callGigachatAPI(req.Text, token)
	if err != nil {
		log.Println("Error calling Gigachat API:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}

func getGigachatToken() (string, error) {
	if gigaToken != "" && time.Now().Before(gigaTokenExpires) {
		return gigaToken, nil
	}

	authKey := os.Getenv("GIGACHAT_AUTH_KEY")
	rquid := uuid.New().String()
	data := []byte("scope=GIGACHAT_API_PERS")

	req, _ := http.NewRequest("POST",
		"https://ngw.devices.sberbank.ru:9443/api/v2/oauth",
		bytes.NewBuffer(data),
	)

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Basic "+authKey)
	req.Header.Set("RqUID", rquid)

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		log.Println("response code:", resp.StatusCode)
		return "", fmt.Errorf("token error %d: %s", resp.StatusCode, string(body))
	}

	var parsed struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", err
	}

	gigaToken = parsed.AccessToken
	gigaTokenExpires = time.Now().Add(time.Duration(parsed.ExpiresIn-60) * time.Second)

	return gigaToken, nil
}

func callGigachatAPI(text, token string) (string, error) {
	body := map[string]any{
		"model": "GigaChat-2",
		"messages": []map[string]string{
			{
				"role": "system",
				"content": `Вы — профессионал по суммаризации текстов.
Ваша задача — создать краткую выжимку Markdown-документа пользователя.
Сохраните все ключевые моменты и структуру, но удалите избыточность и нерелевантные детали.
Выведите только текст суммаризации.
`,
			},
			{
				"role":    "user",
				"content": text,
			},
		},
		"n":                  1,
		"stream":             false,
		"max_tokens":         512,
		"repetition_penalty": 1,
		"update_interval":    0,
	}

	jsonBody, _ := json.Marshal(body)

	req, err := http.NewRequest("POST",
		"https://gigachat.devices.sberbank.ru/api/v1/chat/completions",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		log.Println("Error creating request:", err)
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var parsed struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return "", err
	}

	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}

	return parsed.Choices[0].Message.Content, nil
}

func corsMiddleware() gin.HandlerFunc {
	allowedOrigins := map[string]bool{
		"http://localhost:5173":  true,
		"https://localhost:5173": true,
		fmt.Sprintf("https://%s:%s", os.Getenv("REMOTE_HOST"), os.Getenv("FRONTEND_PORT")): true,
	}

	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if allowedOrigins[origin] {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}
