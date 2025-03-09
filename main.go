package main

import (
	"fmt"
	"os"

	"github.com/moshe5745/localpost/completion" // Completion module
	"github.com/moshe5745/localpost/request"    // Request module

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func main() {
	// Load .env file if present
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error loading .env file: %v\n", err)
	}

	var rootCmd = &cobra.Command{
		Use:   "localpost",
		Short: "A CLI tool to manage and execute HTTP requests",
		Long:  `A tool to save and execute HTTP requests stored in a Git repository.`,
	}

	var sampleNames = []string{"Alice", "Bob", "Charlie", "Dave"}

	var greetCmd = &cobra.Command{
		Use:   "greet [name]",
		Short: "Greet someone",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Hello, %s!\n", args[0])
		},
		ValidArgs: sampleNames,
	}

	var completionCmd = &cobra.Command{
		Use:   "completion",
		Short: "Generate and install completion script automatically",
		Long:  `Automatically detects your shell and installs the completion script in your home directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := completion.InstallCompletion(rootCmd); err != nil {
				fmt.Printf("Error installing completion: %v\n", err)
				os.Exit(1)
			}
		},
	}

	var requestCmd = &cobra.Command{
		Use:   "request [request-file]",
		Short: "Execute an HTTP request from a JSON file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			filePath := args[0]

			// Parse the request
			req, err := request.ParseRequest(filePath)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			// Execute the request
			status, body, err := request.ExecuteRequest(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			fmt.Printf("Response Status: %s\n", status)
			fmt.Printf("Response Body: %s\n", body)
		},
	}

	rootCmd.AddCommand(greetCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(requestCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
