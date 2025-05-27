package exampletool

import (
	"fmt"
	"os"

	"com.terminal-assitant/assistant/internal/llm"
	"com.terminal-assitant/assistant/internal/tools"
)

type CommandHelper struct{}

func CreateCommandHelperTool() tools.Tool {
	return &CommandHelper{}
}

func (t *CommandHelper) Name() string {
	return "command_helper"
}

func (t *CommandHelper) Description() string {
	return `Command Helper is a tool that assists with command-line operations by generating or explaining shell commands based on user input.
Input should be a clear, concise description or question about a command-line task, such as:
- "List all files modified in the last 24 hours"
- "How to find large files in a directory?"
- "Create a backup of my home directory using rsync"
The tool returns suggested commands or explanations to help the user perform the requested task efficiently.`
}

func (t *CommandHelper) Call(input string) (string, error) {
	llmClient := llm.NewOllama(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
	query := fmt.Sprintf(
		"You are a bash command generator. Only output the exact bash command that answers the query, with no explanation, no quotes, no Markdown, and no formatting:\n\n%s",
		input,
	)
	resp, err := llmClient.Invoke(query)
	if err != nil {
		return "", err
	}
	content, ok := resp["message"].(map[string]any)["content"].(string)
	if !ok {
		return "", fmt.Errorf("unexpected response format from LLM")
	}
	return content, nil
}

func (t *CommandHelper) ToolParameters() []tools.ToolParameter {
	return []tools.ToolParameter{
		{
			Name:        "input",
			Description: "The command or query to generate or explain.",
			Type:        "string",
			Required:    true,
		},
	}
}
