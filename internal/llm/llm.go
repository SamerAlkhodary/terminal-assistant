package llm

import "com.terminal-assitant/assistant/internal/tools"

type LLm interface {
	Stream(query string)
	Invoke(query string) (map[string]any, error)
	BindTools(tools []tools.Tool) LLm
	Tools() []tools.Tool
	ToolDescriptions() map[string]string
	ToolNames() []string
}
