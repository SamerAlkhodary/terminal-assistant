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
	"strings"

	"com.terminal-assitant/assistant/internal/llm/model"
	"com.terminal-assitant/assistant/internal/tools"
)

type Ollama struct {
	ollamaUrl   string
	ollamaModel string
	messages    []model.Message
	tools       []tools.Tool
}

// NewOllama creates a new Ollama client with the given URL and model.
func NewOllama(ollamaUrl, ollamaModel string) LLm {
	if ollamaUrl == "" {
		ollamaUrl = "http://localhost:11434"
	}
	if ollamaModel == "" {
		ollamaModel = "llama3.2"
	}
	return &Ollama{
		ollamaUrl:   ollamaUrl,
		ollamaModel: ollamaModel,
		messages:    []model.Message{},
		tools:       []tools.Tool{},
	}
}
func ToolToFunctionDef(tool tools.Tool) map[string]any {
	// Build properties and required fields from Parameters()
	properties := make(map[string]any)
	required := []string{}
	for _, param := range tool.ToolParameters() {
		prop := map[string]any{
			"type":        param.Type,
			"description": param.Description,
		}
		if len(param.Enum) > 0 {
			prop["enum"] = param.Enum
		}
		properties[param.Name] = prop
		if param.Required {
			required = append(required, param.Name)
		}
	}

	// Sanitize tool name
	name := strings.ToLower(strings.ReplaceAll(tool.Name(), " ", "_"))

	return map[string]any{
		"type": "function",
		"function": map[string]any{
			"name":        name,
			"description": tool.Description(),
			"parameters": map[string]any{
				"type":       "object",
				"properties": properties,
				"required":   required,
			},
		},
	}
}

func (ollama *Ollama) createRequest(ollamaUrl, ollamaModel string, stream bool) (*http.Request, error) {
	toolsDefinitions := make([]map[string]any, 0, len(ollama.tools))
	for _, tool := range ollama.tools {
		toolsDefinitions = append(toolsDefinitions, ToolToFunctionDef(tool))
	}
	payload := map[string]any{
		"model":    ollamaModel,
		"messages": ollama.messages,
		"stream":   stream,
		"tools":    toolsDefinitions,
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
func (ollama *Ollama) handleStream(ctx context.Context, body io.ReadCloser) error {
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

func (ollama *Ollama) sendStreamRequest(ctx context.Context, request *http.Request) error {
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
func (ollama *Ollama) sendRequest(ctx context.Context, request *http.Request) (map[string]any, error) {

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("bad status: %s, body: %s", resp.Status, string(b))
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result map[string]any
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, err
	}
	return result, nil

}

func (ollama *Ollama) Invoke(message model.Message) (map[string]any, error) {
	u, err := url.Parse(ollama.ollamaUrl + "/api/chat")
	if err != nil {
		log.Fatalf("Failed to parse Ollama URL: %v", err)
	}
	ollama.messages = append(ollama.messages, message)
	request, err := ollama.createRequest(u.String(), ollama.ollamaModel, false)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	result, err := ollama.sendRequest(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
	jsonResult, _ := json.Marshal(result)
	ollama.messages = append(ollama.messages, model.Message{
		Role:    "assistant",
		Content: string(jsonResult),
	})
	return result, err

}
func (ollama *Ollama) BindTools(tools []tools.Tool) LLm {
	if len(tools) == 0 {
		return ollama
	}
	ollama.tools = tools
	return ollama
}

func (ollama *Ollama) Stream(message model.Message) {
	u, err := url.Parse(ollama.ollamaUrl + "/api/generate")
	if err != nil {
		log.Fatalf("Failed to parse Ollama URL: %v", err)
	}
	ollama.messages = append(ollama.messages, message)
	request, err := ollama.createRequest(u.String(), ollama.ollamaModel, true)
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}
	err = ollama.sendStreamRequest(context.Background(), request)
	if err != nil {
		log.Fatalf("Failed to send request: %v", err)
	}
}
func (o *Ollama) Tools() []tools.Tool {
	return o.tools
}

func (o *Ollama) ToolDescriptions() map[string]string {
	descriptions := make(map[string]string)
	for _, tool := range o.tools {
		descriptions[tool.Name()] = tool.Description()
	}
	return descriptions
}

func (o *Ollama) ToolNames() []string {
	names := make([]string, len(o.tools))
	for i, tool := range o.tools {
		names[i] = tool.Name()
	}
	return names
}
