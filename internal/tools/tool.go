package tools

type Tool interface {
	Name() string
	Description() string
	Call(input string) error
}
