package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Ollama struct {
	ollamaUrl   string
	ollamaModel string
}

// NewOllama creates a new Ollama client with the given URL and model.
func NewOllama(ollamaUrl, ollamaModel string) LLm {
	if ollamaUrl == "" {
		ollamaUrl = "http://localhost:11434"
	}
	if ollamaModel == "" {
		ollamaModel = "llama3.2"
	}
	return Ollama{
		ollamaUrl:   ollamaUrl,
		ollamaModel: ollamaModel,
	}
}

func (ollama Ollama) createRequest(ollamaUrl, ollamaModel, query string) (*http.Request, error) {
	payload := map[string]interface{}{
		"model":  ollamaModel,
		"prompt": query,
		"stream": true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	// Make the POST request
	req, err := http.NewRequest("POST", ollamaUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
func (ollama Ollama) handleStream(ctx context.Context, body io.ReadCloser) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var chunk struct {
			Response string `json:"response"`
		}
		if err := json.Unmarshal(line, &chunk); err == nil {
			fmt.Print(chunk.Response)
		} else {
			// Optionally print the raw line for debugging
			// fmt.Println(string(line))
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}
	fmt.Println()

	return nil
}

func (ollama Ollama) sendRequest(ctx context.Context, request *http.Request) error {
	// Encode the payload to JSON

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bad status: %s, body: %s", resp.Status, string(b))
	}
	return ollama.handleStream(ctx, resp.Body)

}
func (ollama Ollama) Stream(question, command string) {

	ollamaRawUrl := os.Getenv("OLLAMA_HOST")
	ollamaModel := os.Getenv("OLLAMA_MODEL")
	if ollamaRawUrl == "" {
		ollamaRawUrl = "http://localhost:11434"
	}
	if ollamaModel == "" {
		ollamaModel = "llama3.2"
	}
	u, err := url.Parse(ollamaRawUrl + "/api/generate")
	if err != nil {
		log.Fatalf("Failed to parse Ollama URL: %v", err)
	}
	query := ollama.queryBuilder(question, command)
	request, err := ollama.createRequest(u.String(), ollamaModel, query)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	err = ollama.sendRequest(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}

	// Create Ollama API client

	// Print the response
}
func (ollama Ollama) queryBuilder(query string, command string) string {
	if query == "" {
		return ""
	}
	switch command {
	case "command", "c":
		return fmt.Sprintf("You are a bash command generator. Only output the exact bash command that answers the query, with no explanation, no quotes, no Markdown, and no formatting:\n\n%s", query)
	case "question", "q":
		return fmt.Sprintf("You are a command-line, network and dev tools helper only. Give a short,informative and structured answer to the provided question:\n\n%s", query)
	}
	return "Say that you are missing a specifitc command"
}
