package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moshe5745/localpost/install"
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
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var requestFiles []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			matches := methodRegex.FindStringSubmatch(file.Name())
			if len(matches) == 3 {
				method, name := matches[1], matches[2]
				displayName := fmt.Sprintf("'%s %s'", method, strings.TrimSuffix(name, ".json"))
				if strings.HasPrefix(displayName, toComplete) {
					requestFiles = append(requestFiles, displayName)
				}
			}
		}
	}

	return requestFiles, cobra.ShellCompDirectiveNoFileComp
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

	var install = &cobra.Command{
		Use:   "install",
		Short: "Generate and install install script automatically",
		Long:  `Automatically detects your shell and installs the install script in your home directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			if err := completion.InstallCompletion(rootCmd); err != nil {
				fmt.Printf("Error installing install: %v\n", err)
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
			// Split input into method and name (e.g., "GET users")
			parts := strings.SplitN(input, " ", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: invalid input format, expected 'METHOD name' (e.g., GET users)\n")
				os.Exit(1)
			}
			method, name := parts[0], parts[1]

			// Validate method is uppercase (simple check)
			if method != strings.ToUpper(method) {
				fmt.Printf("Error: method must be uppercase (e.g., GET, POST)\n")
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

	rootCmd.AddCommand(install)
	rootCmd.AddCommand(requestCmd)
	rootCmd.AddCommand(newRequestCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
