package agent

import (
	"fmt"
	"strings"

	"com.terminal-assitant/assistant/internal/llm"
	"com.terminal-assitant/assistant/internal/llm/model"
	"com.terminal-assitant/assistant/internal/tools"
)

type Agent struct {
	tools []tools.Tool
	llm   llm.LLm
}

func NewAgent(llm llm.LLm, tools []tools.Tool) *Agent {
	llmWithTools := llm.BindTools(tools)
	llm.Invoke(model.Message{
		Role: "system",
		Content: `You are a tool-using assistant. You can call external tools to retrieve information. Each tool has a name, description, and parameters.

When a user asks a question:
1. Extract the required parameters and call the appropriate tool.
2. When the tool returns a result, evaluate whether it fully answers the user's request.

If the tool result is a shell command or set of commands:
- Return only the raw command(s) as plain text.
- Do NOT add any explanations, descriptions, comments, formatting (like Markdown), or leading phrases.
- Do NOT say "You can use the following command", "Here is the command", or similar.

If the result is not a command, respond with a short and direct answer — no elaboration, examples, or tool references.

If the result is incomplete, use another tool if available or say what is missing. Never guess or invent information.

Respond only when the information is complete. Be concise and strictly follow output rules.`,
	},
	)

	return &Agent{
		llm:   llmWithTools,
		tools: tools,
	}
}

func (a *Agent) Invoke(message model.Message) (string, error) {
	ready := false
	inputMessage := model.Message{
		Role:    "user",
		Content: message.Content,
	}
	response := map[string]any{}
	var err error
	for !ready {
		response, err = a.llm.Invoke(inputMessage)
		if err != nil {
			return "", err
		}
		tool_response, err := a.handleToolCalls(response)
		if err != nil {
			ready = true
		}
		inputMessage.Content = tool_response
		inputMessage.Role = "tool"

	}
	msg, ok := response["message"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("message field missing or wrong type")
	}
	content, ok := msg["content"].(string)
	if !ok {
		return "", fmt.Errorf("content field missing or wrong type")
	}
	return content, nil
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
