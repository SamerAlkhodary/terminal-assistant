package cmd

import (
	"fmt"
	"os"
	"strings"

	"com.terminal-assitant/assistant/internal/llm"
	"com.terminal-assitant/assistant/internal/llmtool"
	"com.terminal-assitant/assistant/internal/tools"
	"github.com/spf13/cobra"
)

var (
	query string
)
var assistant = llm.NewOllama(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
var searchTool = tools.CreateTavilyTool()
var coomandHelperTool = tools.CreateCommandHelperTool()
var toolsList = []tools.Tool{
	searchTool,
	coomandHelperTool,
}
var assistantWithTools, _ = llmtool.Create(assistant, toolsList)

var rootCmd = &cobra.Command{
	Use:   "helper",
	Short: "A CLI tool that help with bash commands",
	Long:  "helper is a simple CLI to help with bash commands.\nSupports flags like --c, q and provides built-in help.",
	Run: func(cmd *cobra.Command, args []string) {

		if query != "" {
			handleInput(query, "q", args)
		} else {
			fmt.Println("Please provide a command or query using --command/-c or --query/-q flags.")
		}
	},
}

func handleInput(input string, flag string, args []string) {
	input += " " + strings.Join(args, " ")
	assistantWithTools.Stream(input)
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
