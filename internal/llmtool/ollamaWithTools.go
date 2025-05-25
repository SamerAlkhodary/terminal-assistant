package llmtool

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"com.terminal-assitant/assistant/internal/llm"
	"com.terminal-assitant/assistant/internal/tools"
)

type OllamaWithTools struct {
	tools []tools.Tool
	llm   llm.LLm
}

func Create(llm llm.LLm, tools []tools.Tool) (LLMWithTools, error) {
	if llm == nil {
		return nil, errors.New("LLM cannot be nil")
	}
	if len(tools) == 0 {
		return nil, errors.New("tools cannot be empty")
	}

	ollamaWithTools := &OllamaWithTools{
		llm:   llm,
		tools: tools,
	}

	return ollamaWithTools, nil
}
func (o *OllamaWithTools) Stream(query string) {
	err := o.invokeTool(query)
	if err != nil {
		fmt.Println("Error invoking tool:", err)
		return
	}
}
func (o *OllamaWithTools) extractToolFromResponse(response string) (string, string) {
	re := regexp.MustCompile(`(?i)TOOL:\s*(.+?)\s*INPUT:\s*(.+)`)
	match := re.FindStringSubmatch(response)
	if len(match) == 3 {
		toolName := match[1]
		inputText := match[2]
		return toolName, inputText
	} else {
		fmt.Println("No match found")
	}
	return "", ""
}
func (o *OllamaWithTools) getTools(query string) (string, error) {
	jsonBytes, err := json.Marshal(o.ToolDescriptions())
	if err != nil {
		panic(err)
	}
	prompt := fmt.Sprintf(`
You are a helpful CLI assistant with these tools:

%s

When answering, you MUST respond in EXACTLY this format, with NO extra text, no missing parts:

TOOL:<tool-name>
INPUT:<input string>

For example:

TOOL:Tavily
INPUT:Weather forecast for Malmö tomorrow

User question: %s
`, string(jsonBytes), query)

	return o.llm.Invoke(prompt)
}
func (o *OllamaWithTools) invokeTool(query string) error {
	resp, err := o.getTools(query)
	if err != nil {
		return fmt.Errorf("error getting tools: %w", err)
	}
	toolName, inputText := o.extractToolFromResponse(resp)

	for _, tool := range o.tools {
		if strings.EqualFold(tool.Name(), toolName) {
			err := tool.Call(inputText)
			if err != nil {
				return fmt.Errorf("error calling tool %s: %w", toolName, err)
			}
			return nil
		}
	}
	return fmt.Errorf("tool %s not found", toolName)

}
func (o *OllamaWithTools) Tools() []tools.Tool {
	return o.tools
}
func (o *OllamaWithTools) ToolDescriptions() map[string]string {
	descriptions := make(map[string]string)
	for _, tool := range o.tools {
		descriptions[tool.Name()] = tool.Description()
	}
	return descriptions
}
func (o *OllamaWithTools) ToolNames() []string {
	names := make([]string, len(o.tools))
	for i, tool := range o.tools {
		names[i] = tool.Name()
	}
	return names
}
