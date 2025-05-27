package agent

import (
	"fmt"
	"strings"

	"com.terminal-assitant/assistant/internal/llm"
	"com.terminal-assitant/assistant/internal/tools"
)

type Agent struct {
	tools []tools.Tool
	llm   llm.LLm
}

func NewAgent(llm llm.LLm, tools []tools.Tool) *Agent {
	llmWithTools := llm.BindTools(tools)
	return &Agent{
		llm:   llmWithTools,
		tools: tools,
	}
}

func (a *Agent) Invoke(input string) (string, error) {
	response, err := a.llm.Invoke(input)
	if err != nil {
		return "", err
	}
	toolResponse, _ := a.handleToolCalls(response)
	return toolResponse, nil
}

func (a *Agent) handleToolCalls(response map[string]any) (string, error) {
	// Extract "message" map
	message, ok := response["message"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("message field missing or wrong type")
	}

	// Extract "tool_calls" slice
	toolCallsRaw, ok := message["tool_calls"]
	if !ok {
		return "", fmt.Errorf("tool_calls field missing")
	}

	toolCalls, ok := toolCallsRaw.([]any)
	if !ok {
		return "", fmt.Errorf("tool_calls field wrong type")
	}

	// Iterate over tool calls, but only handle first valid one
	for _, callRaw := range toolCalls {
		call, ok := callRaw.(map[string]any)
		if !ok {
			continue
		}

		funcMap, ok := call["function"].(map[string]any)
		if !ok {
			continue
		}

		toolName, ok := funcMap["name"].(string)
		if !ok || toolName == "" {
			continue
		}

		args, ok := funcMap["arguments"].(map[string]any)
		if !ok {
			continue
		}

		input, ok := args["input"].(string)
		if !ok {
			continue
		}

		// Find the matching tool by name and call it
		for _, tool := range a.tools {
			if strings.EqualFold(tool.Name(), toolName) {
				result, err := tool.Call(input)
				if err != nil {
					return "", fmt.Errorf("error calling tool %s: %w", toolName, err)
				}
				return result, nil
			}
		}

		return "", fmt.Errorf("tool %s not found", toolName)
	}

	return "", fmt.Errorf("no valid tool call found in response")
}
