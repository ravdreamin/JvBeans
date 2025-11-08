package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type GenerateRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Language string `json:"language,omitempty"`
}

type GenerateResponse struct {
	Code     string `json:"code"`
	Provider string `json:"provider"` // "openai" or "gemini"
}

// Rate limiter
var (
	openaiLimiter = &rateLimiter{
		maxRequests: 50,
		window:      time.Minute,
		requests:    make([]time.Time, 0),
	}
	geminiLimiter = &rateLimiter{
		maxRequests: 100,
		window:      time.Minute,
		requests:    make([]time.Time, 0),
	}
)

type rateLimiter struct {
	sync.Mutex
	maxRequests int
	window      time.Duration
	requests    []time.Time
}

func (rl *rateLimiter) allow() bool {
	rl.Lock()
	defer rl.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests
	validRequests := make([]time.Time, 0)
	for _, t := range rl.requests {
		if t.After(cutoff) {
			validRequests = append(validRequests, t)
		}
	}
	rl.requests = validRequests

	if len(rl.requests) < rl.maxRequests {
		rl.requests = append(rl.requests, now)
		return true
	}
	return false
}

// GenerateCode generates code using AI (OpenAI with Gemini fallback)
func GenerateCode(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	languageContext := ""
	if req.Language != "" {
		languageContext = fmt.Sprintf(" in %s", req.Language)
	}

	systemPrompt := fmt.Sprintf("You are a code generation assistant. Generate clean, well-commented code%s based on the user's request. Return ONLY the code, no explanations or markdown.", languageContext)

	// Try OpenAI first
	if openaiLimiter.allow() {
		code, err := generateWithOpenAI(systemPrompt, req.Prompt)
		if err == nil {
			c.JSON(http.StatusOK, GenerateResponse{
				Code:     code,
				Provider: "openai",
			})
			return
		}
	}

	// Fallback to Gemini
	if geminiLimiter.allow() {
		code, err := generateWithGemini(systemPrompt, req.Prompt)
		if err == nil {
			c.JSON(http.StatusOK, GenerateResponse{
				Code:     code,
				Provider: "gemini",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate code with Gemini: " + err.Error()})
		return
	}

	c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded for both AI providers"})
}

// OpenAI API structures
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []struct {
		Message OpenAIMessage `json:"message"`
	} `json:"choices"`
}

func generateWithOpenAI(systemPrompt, userPrompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	temperature, _ := strconv.ParseFloat(os.Getenv("OPENAI_TEMPERATURE"), 64)
	if temperature == 0 {
		temperature = 0.7
	}

	maxTokens, _ := strconv.Atoi(os.Getenv("OPENAI_MAX_TOKENS"))
	if maxTokens == 0 {
		maxTokens = 2000
	}

	reqBody := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI API error: %s", string(body))
	}

	var openaiResp OpenAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return "", err
	}

	if len(openaiResp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return openaiResp.Choices[0].Message.Content, nil
}

// Gemini API structures
type GeminiRequest struct {
	Contents []GeminiContent `json:"contents"`
}

type GeminiContent struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiPart struct {
	Text string `json:"text"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content GeminiContent `json:"content"`
	} `json:"candidates"`
}

func generateWithGemini(systemPrompt, userPrompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}

	combinedPrompt := fmt.Sprintf("%s\n\nUser request: %s", systemPrompt, userPrompt)

	reqBody := GeminiRequest{
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: combinedPrompt},
				},
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s", model, apiKey)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Gemini API error: %s", string(body))
	}

	var geminiResp GeminiResponse
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}
