package tools

type ToolParameter struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Enum        []string // Optional
}
type Tool interface {
	Name() string
	Description() string
	ToolParameters() []ToolParameter
	Call(input string) (string, error)
}
