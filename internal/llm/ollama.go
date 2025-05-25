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

func (ollama Ollama) createRequest(ollamaUrl, ollamaModel, query string, stream bool) (*http.Request, error) {
	payload := map[string]interface{}{
		"model":  ollamaModel,
		"prompt": query,
		"stream": stream,
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

func (ollama Ollama) sendStreamRequest(ctx context.Context, request *http.Request) error {
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
func (ollama Ollama) sendRequest(ctx context.Context, request *http.Request) (string, error) {

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("bad status: %s, body: %s", resp.Status, string(b))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", err
	}
	responseStr, ok := result["response"].(string)
	if !ok {
		return "", fmt.Errorf("response field is not a string")
	}
	return responseStr, nil

}

func (ollama Ollama) Invoke(query string) (string, error) {
	u, err := url.Parse(ollama.ollamaUrl + "/api/generate")
	if err != nil {
		log.Fatalf("Failed to parse Ollama URL: %v", err)
	}
	request, err := ollama.createRequest(u.String(), ollama.ollamaModel, query, false)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	result, err := ollama.sendRequest(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	return result, err

}

func (ollama Ollama) Stream(query string) {
	u, err := url.Parse(ollama.ollamaUrl + "/api/generate")
	if err != nil {
		log.Fatalf("Failed to parse Ollama URL: %v", err)
	}
	request, err := ollama.createRequest(u.String(), ollama.ollamaModel, query, true)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	err = ollama.sendStreamRequest(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
}
