package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// UPDATED: Added Filename field for better language context
type GenerateRequest struct {
	Prompt   string `json:"prompt" binding:"required"`
	Language string `json:"language,omitempty"`
	Filename string `json:"filename,omitempty"`
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

// UPDATED: Code-only generation with strict prompting and post-processing
func GenerateCode(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Build language context
	languageContext := ""
	if req.Language != "" {
		languageContext = req.Language
	}
	if req.Filename != "" {
		if languageContext != "" {
			languageContext = fmt.Sprintf("%s (file: %s)", languageContext, req.Filename)
		} else {
			languageContext = fmt.Sprintf("file: %s", req.Filename)
		}
	}

	// UPDATED: Strict system prompt for code-only output
	systemPrompt := buildStrictSystemPrompt(languageContext)
	userPrompt := buildCodeOnlyUserPrompt(req.Prompt, languageContext)

	var generatedCode string
	var provider string
	var lastError error

	// Try OpenAI first
	if openaiLimiter.allow() {
		code, err := generateWithOpenAI(systemPrompt, userPrompt)
		if err == nil {
			generatedCode = code
			provider = "openai"
		} else {
			lastError = err
		}
	}

	// Fallback to Gemini if OpenAI failed
	if generatedCode == "" && geminiLimiter.allow() {
		code, err := generateWithGemini(systemPrompt, userPrompt)
		if err == nil {
			generatedCode = code
			provider = "gemini"
		} else {
			lastError = err
		}
	}

	// If both failed, return error
	if generatedCode == "" {
		if lastError != nil && (strings.Contains(lastError.Error(), "API key") || strings.Contains(lastError.Error(), "not set")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"status":  400,
				"code":    "API_KEY_INVALID",
				"message": "No AI provider available. Configure OPENAI_API_KEY or GEMINI_API_KEY in backend .env",
			})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{
			"status":  400,
			"code":    "PROVIDER_UNAVAILABLE",
			"message": "No AI provider available",
		})
		return
	}

	// UPDATED: Post-process to ensure code-only output
	cleanCode := extractCodeOnly(generatedCode)

	// Check if cleaned code is empty
	if strings.TrimSpace(cleanCode) == "" {
		c.JSON(http.StatusOK, GenerateResponse{
			Code:     "",
			Provider: provider,
		})
		return
	}

	c.JSON(http.StatusOK, GenerateResponse{
		Code:     cleanCode,
		Provider: provider,
	})
}

// UPDATED: Build strict system prompt for code-only output
func buildStrictSystemPrompt(languageContext string) string {
	prompt := `You are a code generator.
Output ONLY the final source code for the requested task.
Do NOT include:
- Markdown code fences (no backticks)
- Explanations or prose
- Comments describing the code (unless the user specifically asks for comments)
- Leading or trailing text
- HTML tags like <code> or </code>

Return plain source code text only.`

	if languageContext != "" {
		prompt += fmt.Sprintf("\nTarget: %s", languageContext)
	}

	return prompt
}

// UPDATED: Build user prompt with code-only reminder
func buildCodeOnlyUserPrompt(userRequest, languageContext string) string {
	prompt := userRequest + "\n\nReturn only code. No markdown. No backticks. No explanations."
	if languageContext != "" {
		prompt += fmt.Sprintf("\nTarget language: %s.", languageContext)
	}
	return prompt
}

// UPDATED: Extract code-only from potential markdown/prose response
func extractCodeOnly(text string) string {
	// Strategy 1: If text contains triple-backtick code blocks, extract the first one
	codeBlockRegex := regexp.MustCompile("```[\\w+-]*\\n([\\s\\S]*?)```")
	matches := codeBlockRegex.FindStringSubmatch(text)
	if len(matches) > 1 {
		return strings.TrimSpace(matches[1])
	}

	// Strategy 2: Remove common prose patterns
	lines := strings.Split(text, "\n")
	var codeLines []string
	inCode := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines at the beginning
		if !inCode && trimmed == "" {
			continue
		}

		// Skip common prose markers
		if !inCode && (strings.HasPrefix(trimmed, "Here is") ||
			strings.HasPrefix(trimmed, "Here's") ||
			strings.HasPrefix(trimmed, "Explanation:") ||
			strings.HasPrefix(trimmed, "This code") ||
			trimmed == "```") {
			continue
		}

		// Found actual code
		inCode = true
		codeLines = append(codeLines, line)
	}

	result := strings.Join(codeLines, "\n")
	return strings.TrimSpace(result)
}

// OpenAI API structures
type OpenAIRequest struct {
	Model       string          `json:"model"`
	Messages    []OpenAIMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
	TopP        float64         `json:"top_p"`
	Stop        []string        `json:"stop,omitempty"`
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

// UPDATED: OpenAI with strict parameters and stop sequences
func generateWithOpenAI(systemPrompt, userPrompt string) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY not set")
	}

	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}

	// UPDATED: Lower temperature for more deterministic code
	temperature := 0.2
	if tempStr := os.Getenv("OPENAI_TEMPERATURE"); tempStr != "" {
		if parsed, err := strconv.ParseFloat(tempStr, 64); err == nil {
			temperature = parsed
		}
	}

	maxTokens := 800
	if tokensStr := os.Getenv("OPENAI_MAX_TOKENS"); tokensStr != "" {
		if parsed, err := strconv.Atoi(tokensStr); err == nil {
			maxTokens = parsed
		}
	}

	reqBody := OpenAIRequest{
		Model: model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
		Temperature: temperature,
		MaxTokens:   maxTokens,
		TopP:        1.0,
		Stop:        []string{"```", "<code>", "</code>"}, // UPDATED: Stop sequences
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
	Contents         []GeminiContent   `json:"contents"`
	SystemInstruction *GeminiInstruction `json:"systemInstruction,omitempty"`
	GenerationConfig  *GeminiGenConfig   `json:"generationConfig,omitempty"`
}

type GeminiInstruction struct {
	Parts []GeminiPart `json:"parts"`
}

type GeminiGenConfig struct {
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"maxOutputTokens,omitempty"`
	TopP        float64 `json:"topP"`
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

// UPDATED: Gemini with system instruction and strict parameters
func generateWithGemini(systemPrompt, userPrompt string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not set")
	}

	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}

	// UPDATED: Use system instruction for Gemini
	reqBody := GeminiRequest{
		SystemInstruction: &GeminiInstruction{
			Parts: []GeminiPart{
				{Text: systemPrompt},
			},
		},
		Contents: []GeminiContent{
			{
				Parts: []GeminiPart{
					{Text: userPrompt},
				},
			},
		},
		GenerationConfig: &GeminiGenConfig{
			Temperature: 0.2,
			MaxTokens:   800,
			TopP:        1.0,
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
