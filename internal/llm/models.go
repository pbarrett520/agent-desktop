package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ModelInfo represents a model returned by the /v1/models endpoint.
type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	OwnedBy string `json:"owned_by"`
}

// modelsResponse is the response from GET /v1/models.
type modelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

// FetchModels queries GET {endpoint}/models and returns available chat models.
// It filters out embedding models (IDs containing "embed").
// The endpoint should be the base URL with /v1 suffix, e.g. "http://localhost:1234/v1".
// apiKey can be empty for local endpoints.
func FetchModels(ctx context.Context, endpoint string, apiKey string) ([]ModelInfo, error) {
	endpoint = strings.TrimSuffix(endpoint, "/")
	url := fmt.Sprintf("%s/models", endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var modelsResp modelsResponse
	if err := json.Unmarshal(body, &modelsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Filter out embedding models
	var chatModels []ModelInfo
	for _, m := range modelsResp.Data {
		idLower := strings.ToLower(m.ID)
		if strings.Contains(idLower, "embed") {
			continue
		}
		chatModels = append(chatModels, m)
	}

	return chatModels, nil
}

// CheckEndpointHealth checks if an OpenAI-compatible endpoint is reachable
// by hitting GET /v1/models. Does NOT require a model to be configured
// and does NOT trigger JIT model loading.
func CheckEndpointHealth(ctx context.Context, endpoint string, apiKey string) (bool, string) {
	models, err := FetchModels(ctx, endpoint, apiKey)
	if err != nil {
		return false, "Endpoint unreachable: " + err.Error()
	}
	return true, fmt.Sprintf("Connected. %d model(s) available.", len(models))
}
