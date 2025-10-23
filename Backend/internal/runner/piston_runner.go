package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// httpClient is a shared HTTP client with a timeout to prevent hanging requests.
var httpClient = &http.Client{
	Timeout: 10 * time.Second, // A reasonable timeout for code execution
}

// getFileExtension returns a common file extension for a given language.
// This is a simplified mapping and might need to be expanded for more languages.
func getFileExtension(language string) string {
	switch language {
	case "java":
		return "java"
	case "python":
		return "py"
	case "javascript":
		return "js"
	default:
		return "txt" // Fallback for unknown languages
	}
}

// PistonRequest is the structure for the Piston API.
type PistonRequest struct {
	Language string              `json:"language"`
	Version  string              `json:"version"`
	Files    []map[string]string `json:"files"`
}

// ExecuteCode sends code to the Piston API and returns the result
func ExecuteCode(language, code string) (map[string]interface{}, error) {
	filename := fmt.Sprintf("Main.%s", getFileExtension(language))

	payload := PistonRequest{
		Language: language,
		Version:  "latest",
		Files: []map[string]string{
			{"name": filename, "content": code},
		},
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Piston request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://emkc.org/api/v2/piston/execute", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req) // Use the shared HTTP client with timeout
	if err != nil {
		return nil, fmt.Errorf("failed to send request to Piston API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Attempt to read Piston's error response for more details
		var errorResult map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errorResult); decodeErr == nil {
			return nil, fmt.Errorf("Piston API returned status %d: %v", resp.StatusCode, errorResult)
		}
		return nil, fmt.Errorf("Piston API returned status %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode Piston API response: %w", err)
	}

	return result, nil
}
