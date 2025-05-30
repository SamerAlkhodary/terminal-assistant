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
	response := model.Response{}
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
	return response.Message.Content, nil
}

func (a *Agent) handleToolCalls(response model.Response) (string, error) {
	// Extract "message" map

	// Iterate over tool calls, but only handle first valid one
	for _, call := range response.Message.ToolCalls {

		// Find the matching tool by name and call it
		for _, tool := range a.tools {
			if strings.EqualFold(tool.Name(), call.Function.Name) {
				result, err := tool.Call(call.Function.Arguments.Input)
				if err != nil {
					return "", fmt.Errorf("error calling tool %s: %w", call.Function.Name, err)
				}
				return result, nil
			}
		}

	}

	return "", fmt.Errorf("no valid tool call found in response")
}
