package llmtool

import "com.terminal-assitant/assistant/internal/tools"

type LLMWithTools interface {
	Stream(query string)
	Tools() []tools.Tool
	ToolDescriptions() map[string]string
	ToolNames() []string
}
