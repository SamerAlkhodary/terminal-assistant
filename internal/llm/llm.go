package llm

import (
	"com.terminal-assitant/assistant/internal/llm/model"
	"com.terminal-assitant/assistant/internal/tools"
)

type LLm interface {
	Stream(model.Message)
	Invoke(model.Message) (model.Response, error)
	BindTools(tools []tools.Tool) LLm
	Tools() []tools.Tool
	ToolDescriptions() map[string]string
	ToolNames() []string
}
