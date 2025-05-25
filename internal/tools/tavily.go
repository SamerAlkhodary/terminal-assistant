package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Tavily struct {
	apiKey string
}

func CreateTavilyTool() Tool {
	apiKey, _ := getAPIKey()
	return &Tavily{
		apiKey: apiKey,
	}
}
func (t *Tavily) Name() string {
	return "Tavily"
}
func (t *Tavily) Description() string {
	return `Tavily is a tool for answering questions that require up-to-date or web-based information.

Use it whenever the user asks about:
- Current events
- Weather
- Recent developments
- Factual web search queries

Input should be a short search phrase or topic, for example:
- Latest news on AI advancements
- Weather forecast for New York tomorrow
- Top restaurants in San Francisco

The tool returns relevant, concise information gathered from trusted online sources.`
}

func (t *Tavily) Call(input string) error {
	// Here you would implement the logic to call the Tavily API using t.ApiKey
	// For now, we will return a placeholder response
	payload, err := t.buildSearchPayload(input)
	if err != nil {
		fmt.Println("Failed to build payload:", err)
		return err
	}

	response, err := t.sendTavilyRequest(payload)
	if err != nil {
		fmt.Println("Failed to send request:", err)
		return err
	}

	t.handleTavilyResponse(response)
	return nil
}

// Helper to get the Tavily API key from the environment
func getAPIKey() (string, error) {
	apiKey := os.Getenv("TAVILY_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("TAVILY_API_KEY environment variable not set")
	}
	return apiKey, nil
}

// Helper to build the search request payload
func (t *Tavily) buildSearchPayload(query string) ([]byte, error) {
	payload := map[string]interface{}{
		"query":                      query,
		"include_answer":             true,
		"include_raw_content":        false,
		"include_images":             false,
		"include_image_descriptions": false,
		"max_results":                1,
	}
	return json.Marshal(payload)
}

// Helper to send the HTTP request to Tavily
func (t *Tavily) sendTavilyRequest(payload []byte) ([]byte, error) {
	url := "https://api.tavily.com/search"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

// Helper to handle and print the response
func (t *Tavily) handleTavilyResponse(response []byte) {
	var result map[string]interface{}
	err := json.Unmarshal(response, &result)
	if err != nil {
		fmt.Println("Failed to parse Tavily response:", err)
	}
	// Type assertion to get the "answer" as a string
	if answer, ok := result["answer"].(string); ok {
		fmt.Println(answer)

	}

}
