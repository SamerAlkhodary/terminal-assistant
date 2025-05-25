package tools

import (
	"fmt"
	"os"

	"com.terminal-assitant/assistant/internal/llm"
)

type CommandHelper struct {
	ApiKey string
}

func CreateCommandHelperTool() Tool {
	return &CommandHelper{}
}
func (t *CommandHelper) Name() string {
	return "Command Helper"
}
func (t *CommandHelper) Description() string {
	return `Command Helper is a tool that assists with command-line operations by generating or explaining shell commands based on user input. 
		Input should be a clear, concise description or question about a command-line task, such as:
		- "List all files modified in the last 24 hours"
		- "How to find large files in a directory?"
		- "Create a backup of my home directory using rsync"
		The tool returns suggested commands or explanations to help the user perform the requested task efficiently.`
}
func (t *CommandHelper) Call(input string) error {
	llm := llm.NewOllama(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
	query := fmt.Sprintf("You are a bash command generator. Only output the exact bash command that answers the query, with no explanation, no quotes, no Markdown, and no formatting:\n\n%s", input)
	llm.Stream(query)
	return nil
}
