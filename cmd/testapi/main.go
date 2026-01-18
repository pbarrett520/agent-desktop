// Live test for OpenAI-compatible API connection
// Run with: go run ./cmd/testapi
package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	// Support both old Azure-style and new generic env vars
	endpoint := os.Getenv("LLM_ENDPOINT")
	apiKey := os.Getenv("LLM_API_KEY")
	model := os.Getenv("LLM_MODEL")

	// Fallback to OpenAI-specific vars
	if endpoint == "" {
		endpoint = os.Getenv("OPENAI_API_BASE")
		if endpoint == "" {
			endpoint = "https://api.openai.com/v1"
		}
	}
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if model == "" {
		model = os.Getenv("OPENAI_MODEL")
		if model == "" {
			model = "gpt-4o"
		}
	}

	// Check for required values
	if apiKey == "" {
		fmt.Println("=== OpenAI-Compatible API Connection Test ===")
		fmt.Println()
		fmt.Println("Missing required environment variables!")
		fmt.Println()
		fmt.Println("Create a .env file in the project root with:")
		fmt.Println("  LLM_ENDPOINT=https://api.openai.com/v1  (or your provider)")
		fmt.Println("  LLM_API_KEY=your-api-key")
		fmt.Println("  LLM_MODEL=gpt-4o  (or your model)")
		fmt.Println()
		fmt.Println("Supported providers:")
		fmt.Println("  - OpenAI: https://api.openai.com/v1")
		fmt.Println("  - LM Studio: http://localhost:1234/v1")
		fmt.Println("  - OpenRouter: https://openrouter.ai/api/v1")
		fmt.Println("  - Any OpenAI-compatible API")
		fmt.Println()
		fmt.Println("Or set these as environment variables.")
		os.Exit(1)
	}

	fmt.Println("=== OpenAI-Compatible API Connection Test ===")
	fmt.Printf("Endpoint: %s\n", endpoint)
	fmt.Printf("Model:    %s\n", model)
	if len(apiKey) > 8 {
		fmt.Printf("API Key:  %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
	} else {
		fmt.Printf("API Key:  ***\n")
	}
	fmt.Println()

	// Clean endpoint
	endpoint = strings.TrimSuffix(endpoint, "/")

	// Test the connection
	fmt.Println("Testing connection...")
	success, err := testConnection(endpoint, apiKey, model)
	if success {
		fmt.Printf("✅ SUCCESS - Connected to %s\n", endpoint)
		fmt.Println()

		// Try a chat completion with tools
		fmt.Println("Testing chat completion with tools...")
		testChatCompletion(endpoint, apiKey, model)
	} else {
		fmt.Printf("❌ FAILED: %v\n", err)
		fmt.Println()

		// Try direct HTTP to see what's happening
		fmt.Println("Testing direct HTTP request...")
		testDirectHTTP(endpoint, apiKey, model)
	}
}

func testConnection(endpoint, apiKey, model string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/chat/completions", endpoint)

	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "user", "content": "Say 'hello' and nothing else."},
		},
		"max_tokens": 10,
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, err
	}

	// Check for response content
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := msg["content"].(string); ok {
					fmt.Printf("Response: %s\n", content)
					return true, nil
				}
			}
		}
	}

	return false, fmt.Errorf("unexpected response format")
}

func testChatCompletion(endpoint, apiKey, model string) {
	fmt.Println("\n=== Testing with Tools ===")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	url := fmt.Sprintf("%s/chat/completions", endpoint)

	body := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant. Use tools when appropriate."},
			{"role": "user", "content": "What time is it?"},
		},
		"tools": []map[string]interface{}{
			{
				"type": "function",
				"function": map[string]interface{}{
					"name":        "get_current_time",
					"description": "Get the current time",
					"parameters":  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
				},
			},
		},
	}

	bodyBytes, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(string(bodyBytes)))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Printf("Error decoding response: %v\n", err)
		return
	}

	// Pretty print the response
	prettyJSON, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Response:\n%s\n", string(prettyJSON))
}

func testDirectHTTP(endpoint, apiKey, model string) {
	url := fmt.Sprintf("%s/chat/completions", endpoint)

	body := fmt.Sprintf(`{"model":"%s","messages":[{"role":"user","content":"hello"}],"max_tokens":10}`, model)

	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	fmt.Printf("Request URL: %s\n", url)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	fmt.Printf("Response Status: %s\n", resp.Status)

	// Read response body
	scanner := bufio.NewScanner(resp.Body)
	var responseBody strings.Builder
	for scanner.Scan() {
		responseBody.WriteString(scanner.Text())
	}

	fmt.Printf("Response Body: %s\n", responseBody.String())
}
