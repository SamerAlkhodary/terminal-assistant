package llm

type LLm interface {
	Stream(query string)
	Invoke(query string) (string, error)
}
