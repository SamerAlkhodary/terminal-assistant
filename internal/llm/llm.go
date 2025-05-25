package llm

type LLm interface {
	Stream(query string, flag string)
}
