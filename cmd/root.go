package cmd

import (
	"fmt"
	"os"
	"strings"

	"com.terminal-assitant/assistant/internal/agent"
	"com.terminal-assitant/assistant/internal/exampletool"
	"com.terminal-assitant/assistant/internal/llm"
	"com.terminal-assitant/assistant/internal/llm/model"
	"com.terminal-assitant/assistant/internal/tools"
	"github.com/spf13/cobra"
)

var (
	query string
)

var (
	assistant         = llm.NewOllama(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
	searchTool        = exampletool.CreateTavilyTool()
	commandHelperTool = exampletool.CreateCommandHelperTool()
	toolsList         = []tools.Tool{searchTool, commandHelperTool}
	assistantAgent    = agent.NewAgent(assistant, toolsList)
)

var rootCmd = &cobra.Command{
	Use:   "helper",
	Short: "A CLI tool that helps with bash commands",
	Long:  "helper is a simple CLI to help with bash commands.\nSupports flags like --c, -q and provides built-in help.",
	Run: func(cmd *cobra.Command, args []string) {
		if query != "" {
			handleInput(query, args)
		} else {
			fmt.Println("Please provide a command or query using --command/-c or --query/-q flags.")
		}
	},
}

func handleInput(input string, args []string) {
	fullInput := strings.TrimSpace(input + " " + strings.Join(args, " "))
	message := model.Message{
		Role:    "user",
		Content: fullInput,
	}
	response, err := assistantAgent.Invoke(message)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("response:", response)
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&query, "query", "q", "", "Ask helper for a query")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
