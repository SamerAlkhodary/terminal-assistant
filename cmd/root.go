package cmd

import (
	"fmt"
	"os"
	"strings"

	"com.terminal-assitant/assistant/internal/llm"
	"github.com/spf13/cobra"
)

var (
	command  string
	question string
)
var assistant = llm.NewOllama(os.Getenv("OLLAMA_URL"), os.Getenv("OLLAMA_MODEL"))
var rootCmd = &cobra.Command{
	Use:   "helper",
	Short: "A CLI tool that help with bash commands",
	Long:  "helper is a simple CLI to help with bash commands.\nSupports flags like --c, q and provides built-in help.",
	Run: func(cmd *cobra.Command, args []string) {
		if command != "" {
			handleInput(command, "c", args)
		} else if question != "" {
			handleInput(question, "q", args)
		} else {
			fmt.Println("Please provide a command or question using --command/-c or --question/-q flags.")
		}
	},
}

func handleInput(input string, flag string, args []string) {
	input += " " + strings.Join(args, " ")
	assistant.Stream(input, flag)
}

func init() {
	// Custom flag: --question / -q
	rootCmd.PersistentFlags().StringVarP(&command, "command", "c", "", "Ask helper to generate a bash command for your question")
	rootCmd.PersistentFlags().StringVarP(&question, "question", "q", "", "Ask helper for a general question")

}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
