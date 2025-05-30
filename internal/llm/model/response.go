package model

type Response struct {
	Model     string  `json:"model"`
	CreatedAt string  `json:"created_at"`
	Message   Message `json:"message"`
}

type Message struct {
	Role      string      `json:"role"`
	Content   string      `json:"content"`
	ToolCalls []*ToolCall `json:"tool_calls,omitempty"`
}
type ToolCall struct {
	Function FunctionCall `json:"function,omitempty"`
}
type FunctionCall struct {
	Name      string   `json:"name"`
	Arguments Argument `json:"arguments"`
}
type Argument struct {
	Input string `json:"input"`
}
