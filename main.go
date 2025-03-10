package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moshe5745/localpost/completion"
	"github.com/moshe5745/localpost/request"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var methodRegex = regexp.MustCompile(`^\[([A-Z]+)\](.+)$`)

func getRequestFiles(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	requestsDir := os.Getenv("LOCALPOST_REQUESTS_DIR")
	if requestsDir == "" {
		requestsDir = "requests"
	}

	files, err := os.ReadDir(requestsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "DEBUG: Error reading %s: %v\n", requestsDir, err)
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var requestFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			matches := methodRegex.FindStringSubmatch(file.Name())
			if len(matches) == 3 {
				method, name := matches[1], matches[2]
				// Remove brackets from display name
				displayName := method + strings.TrimSuffix(name, ".json")
				if strings.HasPrefix(displayName, toComplete) {
					requestFiles = append(requestFiles, displayName)
				}
			}
		}
	}

	return requestFiles, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

func main() {
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
			input := args[0]
			// Extract method and name from input (e.g., "GETusers")
			// Assume method is all uppercase until a lowercase letter or end
			var method, name string
			for i, r := range input {
				if r >= 'A' && r <= 'Z' {
					method += string(r)
				} else {
					name = input[i:]
					break
				}
			}
			if method == "" || name == "" {
				fmt.Printf("Error: invalid input format, expected METHODname (e.g., GETusers)\n")
				os.Exit(1)
			}

			// Map back to filename with brackets
			requestsDir := os.Getenv("LOCALPOST_REQUESTS_DIR")
			if requestsDir == "" {
				requestsDir = "requests"
			}
			filePath := filepath.Join(requestsDir, fmt.Sprintf("[%s]%s.json", method, name))

			req, err := request.ParseRequest(filePath)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}

			req.Method = method
			req.URL = fmt.Sprintf("{{BASE_URL}}/%s", name)

			status, body, err := request.ExecuteRequest(req)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Response Status: %s\n", status)
			fmt.Printf("Response Body: %s\n", body)
		},
		ValidArgsFunction: getRequestFiles,
	}

	var newRequestCmd = &cobra.Command{
		Use:   "new-request [method] [name]",
		Short: "Create a new request JSON file",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			method, name := strings.ToUpper(args[0]), args[1]
			requestsDir := os.Getenv("LOCALPOST_REQUESTS_DIR")
			if requestsDir == "" {
				requestsDir = "requests"
			}
			filePath := filepath.Join(requestsDir, fmt.Sprintf("[%s]%s.json", method, name))
			if err := os.MkdirAll(requestsDir, 0755); err != nil {
				fmt.Printf("Error creating requests directory: %v\n", err)
				os.Exit(1)
			}
			req := request.Request{
				Headers: map[string]string{"Accept": "application/json"},
				Body:    "",
			}
			data, err := json.MarshalIndent(req, "", "  ")
			if err != nil {
				fmt.Printf("Error marshaling request: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(filePath, data, 0644); err != nil {
				fmt.Printf("Error writing request file: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Created new request file: %s\n", filePath)
		},
	}

	rootCmd.AddCommand(greetCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(requestCmd)
	rootCmd.AddCommand(newRequestCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
