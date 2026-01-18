// Live test for Azure OpenAI API connection
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
	openai "github.com/sashabaranov/go-openai"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using environment variables")
	}

	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	apiKey := os.Getenv("AZURE_OPENAI_KEY")
	deployment := os.Getenv("AZURE_OPENAI_DEPLOYMENT")
	model := os.Getenv("AZURE_OPENAI_MODEL")

	// Check for required values
	if endpoint == "" || apiKey == "" || deployment == "" {
		fmt.Println("=== Azure OpenAI Connection Test ===")
		fmt.Println()
		fmt.Println("Missing required environment variables!")
		fmt.Println()
		fmt.Println("Create a .env file in the project root with:")
		fmt.Println("  AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com")
		fmt.Println("  AZURE_OPENAI_KEY=your-api-key")
		fmt.Println("  AZURE_OPENAI_DEPLOYMENT=your-deployment-name")
		fmt.Println("  AZURE_OPENAI_MODEL=gpt-4")
		fmt.Println()
		fmt.Println("Or set these as environment variables.")
		os.Exit(1)
	}

	fmt.Println("=== Azure OpenAI Connection Test ===")
	fmt.Printf("Endpoint:   %s\n", endpoint)
	fmt.Printf("Deployment: %s\n", deployment)
	fmt.Printf("Model:      %s\n", model)
	if len(apiKey) > 8 {
		fmt.Printf("API Key:    %s...%s\n", apiKey[:4], apiKey[len(apiKey)-4:])
	} else {
		fmt.Printf("API Key:    ***\n")
	}
	fmt.Println()

	// Clean endpoint
	endpoint = strings.TrimSuffix(endpoint, "/")

	// Test different API versions
	apiVersions := []string{
		"2024-10-21",
		"2024-08-01-preview",
		"2024-06-01",
		"2024-02-15-preview",
		"2023-12-01-preview",
		"2023-05-15",
	}

	fmt.Println("Testing API versions...")
	fmt.Println()

	for _, apiVersion := range apiVersions {
		fmt.Printf("Testing API version: %s\n", apiVersion)
		success, err := testConnection(endpoint, apiKey, deployment, apiVersion)
		if success {
			fmt.Printf("  ✅ SUCCESS with API version %s\n", apiVersion)
			fmt.Println()

			// Try a simple chat completion
			fmt.Println("Testing chat completion...")
			testChatCompletion(endpoint, apiKey, deployment, apiVersion)
			return
		} else {
			fmt.Printf("  ❌ FAILED: %v\n", err)
		}
		fmt.Println()
	}

	// If all API versions failed, let's try direct HTTP to see what's happening
	fmt.Println("All API versions failed. Testing direct HTTP request...")
	testDirectHTTP(endpoint, apiKey, deployment)
}

func testConnection(endpoint, apiKey, deployment, apiVersion string) (bool, error) {
	config := openai.DefaultAzureConfig(apiKey, endpoint)
	config.APIVersion = apiVersion

	client := openai.NewClientWithConfig(config)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: deployment,
		Messages: []openai.ChatCompletionMessage{
			{Role: "user", Content: "Say 'hello' and nothing else."},
		},
		MaxTokens: 10,
	})

	if err != nil {
		return false, err
	}

	if len(resp.Choices) > 0 {
		fmt.Printf("  Response: %s\n", resp.Choices[0].Message.Content)
		return true, nil
	}

	return false, fmt.Errorf("no choices in response")
}

func testChatCompletion(endpoint, apiKey, deployment, apiVersion string) {
	fmt.Println("\n=== Testing with custom AzureClient ===")
	
	// Test using our custom client
	cfg := &config{
		endpoint:   endpoint,
		apiKey:     apiKey,
		deployment: deployment,
		apiVersion: apiVersion,
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Make direct HTTP request with tools
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=%s",
		endpoint, deployment, apiVersion)

	body := map[string]interface{}{
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
	req.Header.Set("api-key", apiKey)

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
	
	_ = cfg // suppress unused warning
}

type config struct {
	endpoint   string
	apiKey     string
	deployment string
	apiVersion string
}

func testDirectHTTP(endpoint, apiKey, deployment string) {
	// Try to make a direct HTTP request to see the raw response
	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-10-21", endpoint, deployment)
	
	body := `{"messages":[{"role":"user","content":"hello"}],"max_tokens":10}`
	
	req, err := http.NewRequest("POST", url, strings.NewReader(body))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)
	
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
