package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

const pistonAPI = "https://emkc.org/api/v2/piston"

type RunRequest struct {
	Language string `json:"language" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type PistonRequest struct {
	Language string   `json:"language"`
	Version  string   `json:"version"`
	Files    []File   `json:"files"`
	Stdin    string   `json:"stdin"`
	Args     []string `json:"args"`
}

type File struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type PistonResponse struct {
	Run struct {
		Stdout string `json:"stdout"`
		Stderr string `json:"stderr"`
		Code   int    `json:"code"`
		Output string `json:"output"`
	} `json:"run"`
}

// RunCode executes code using Piston API
func RunCode(c *gin.Context) {
	var req RunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Map frontend language names to Piston language names
	pistonLang := mapLanguageToPiston(req.Language)
	filename := getFilenameForLanguage(req.Language)

	pistonReq := PistonRequest{
		Language: pistonLang,
		Version:  "*", // Use latest version
		Files: []File{
			{
				Name:    filename,
				Content: req.Code,
			},
		},
		Stdin: "",
		Args:  []string{},
	}

	// Make request to Piston API
	jsonData, err := json.Marshal(pistonReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encode request"})
		return
	}

	resp, err := http.Post(pistonAPI+"/execute", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to execute code"})
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read response"})
		return
	}

	var pistonResp PistonResponse
	if err := json.Unmarshal(body, &pistonResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	// Return result
	c.JSON(http.StatusOK, gin.H{
		"stdout": pistonResp.Run.Stdout,
		"stderr": pistonResp.Run.Stderr,
		"code":   pistonResp.Run.Code,
		"output": pistonResp.Run.Output,
	})
}

// mapLanguageToPiston maps our language names to Piston's expected names
func mapLanguageToPiston(language string) string {
	mapping := map[string]string{
		"javascript": "javascript",
		"python":     "python",
		"java":       "java",
		"c":          "c",
		"cpp":        "cpp",
		"go":         "go",
		"typescript": "typescript",
		"rust":       "rust",
	}

	if pistonLang, exists := mapping[language]; exists {
		return pistonLang
	}
	return "javascript" // default
}

// getFilenameForLanguage returns appropriate filename for the language
func getFilenameForLanguage(language string) string {
	mapping := map[string]string{
		"javascript": "main.js",
		"python":     "main.py",
		"java":       "Main.java",
		"c":          "main.c",
		"cpp":        "main.cpp",
		"go":         "main.go",
		"typescript": "main.ts",
		"rust":       "main.rs",
	}

	if filename, exists := mapping[language]; exists {
		return filename
	}
	return "main.js" // default
}
