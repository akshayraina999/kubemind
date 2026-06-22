package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles communications with the local Ollama daemon.
type Client struct {
	BaseURL string
	Model   string
}

// OllamaRequest maps the strict request parameters expected by Ollama's /api/generate endpoint.
type OllamaRequest struct {
	Model   string `json:"model"`
	Prompt  string `json:"prompt"`
	Stream  bool   `json:"stream"`
	Format  string `json:"format"` // Forces JSON mode if set to "json"
}

// OllamaResponse unpacks the raw metadata text blocks sent back by the engine.
type OllamaResponse struct {
	Response string `json:"response"`
}

// NewClient creates a configured instance of our AI pipeline handler.
func NewClient(url, model string) *Client {
	return &Client{
		BaseURL: url,
		Model:   model,
	}
}

// Generate takes our structured cluster context strings and returns the AI's diagnosis.
func (c *Client) Generate(prompt string, forceJSON bool) (string, error) {
	apiURL := fmt.Sprintf("%s/api/generate", c.BaseURL)

	reqPayload := OllamaRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
	}

	if forceJSON {
		reqPayload.Format = "json"
	}

	// Convert our Go struct completely into an in-memory JSON payload byte array
	jsonBytes, err := json.Marshal(reqPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Establish a native HTTP client with a strict timeout boundary
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Post(apiURL, "application/json", bytes.NewBuffer(jsonBytes))
	if err != nil {
		return "", fmt.Errorf("ollama connection failed: %w", err)
	}
	defer resp.Body.Close() // Ensures network sockets drain and close cleanly

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("ollama returned non-200 status: %d", resp.StatusCode)
	}

	// Read the streaming body channel data fields
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal ollama payload: %w", err)
	}

	return ollamaResp.Response, nil
}