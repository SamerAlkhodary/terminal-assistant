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
		Content: `
You are a tool-using assistant. You can call external tools to retrieve information. Each tool has a name, description, and parameters.

When a user asks a question:
1. Extract the required parameters and the tools needed to handle the question.
2. If the question requires tools, call the correct tools with the extracted parameters.
3. When the tool returns a result, evaluate whether it fully answers the user's request.
IMPORTANT:
	If the tool call result is a command or list of commands then:
	- You MUST return only the exact string of the shell command(s).
	- Do NOT add explanations, formatting, or any extra words.
	- NEVER rewrite, summarize, or interpret the command.
	- Example: if the tool returns "ls -a", you must respond with exactly: ls -a
	If the result is not a command, respond with a short and direct answer using the result you recevied from the tool.

Respond only when the information is complete and with the tools' responses. Be concise and strictly follow output rules.`,
	},
	)

	return &Agent{
		llm:   llmWithTools,
		tools: tools,
	}
}

func (a *Agent) Invoke(message model.Message) (string, error) {
	inputMessage := model.Message{
		Role:    "user",
		Content: message.Content,
	}
	response := model.Response{}
	var err error
	maxSteps := 5
	for range maxSteps {
		response, err = a.llm.Invoke(inputMessage)
		if err != nil {
			return "", err
		}
		if len(response.Message.ToolCalls) == 0 {
			break
		}
		toolResponse, _ := a.handleToolCalls(response)
		fmt.Println("toolResponse:", fmt.Sprint(toolResponse))

		inputMessage = model.Message{
			Role:    "tool",
			Content: toolResponse,
		}

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
				return "tool response:" + result, nil
			}
		}

	}

	return "", fmt.Errorf("no valid tool call found in response")
}
