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
	"github.com/joho/godotenv"
)

var gigaToken string
var gigaTokenExpires time.Time

func main() {
	_ = godotenv.Load()

	r := gin.Default()
	r.Use(corsMiddleware())

	r.POST("/api/gigachat/summarize", summarizeHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	fmt.Println("Gigachat Proxy running on :" + port)
	log.Fatal(r.Run(":" + port))
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
				"content": `You are a professional summarization assistant. 
Your task is to create a concise, factual summary of the user's Markdown document. 
Preserve all key points and structure, but remove redundancy and irrelevant details.
Output only the summary text.
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
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}
		c.Next()
	}
}
