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
	"gopkg.in/yaml.v3"
)

var methodRegex = regexp.MustCompile(`^\[([A-Z]+)\](.+)$`)
var env string

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
	if len(files) == 0 {
		requestFiles = append(requestFiles, "No requests found in "+requestsDir)
		return requestFiles, cobra.ShellCompDirectiveError | cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			matches := methodRegex.FindStringSubmatch(file.Name())
			if len(matches) == 3 {
				method, name := matches[1], matches[2]
				displayName := fmt.Sprintf("%s %s", method, strings.TrimSuffix(name, ".json"))
				if strings.HasPrefix(displayName, toComplete) {
					requestFiles = append(requestFiles, displayName)
				}
			}
		}
	}

	return requestFiles, cobra.ShellCompDirectiveNoSpace | cobra.ShellCompDirectiveNoFileComp
}

type Config struct {
	DefaultEnv string                       `yaml:"default_env"`
	Envs       map[string]map[string]string `yaml:"envs"`
}

func loadConfig() (string, error) {
	configFilePath := filepath.Join(os.Getenv("HOME"), ".localpost")
	config := Config{
		DefaultEnv: "dev",
		Envs:       make(map[string]map[string]string),
	}

	data, err := os.ReadFile(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "dev", nil // File doesn’t exist, use default
		}
		return "", fmt.Errorf("error reading .localpost: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid .localpost content: %s\n", string(data))
		return "", fmt.Errorf("error parsing .localpost: %v", err)
	}

	// Load env vars for the current env
	if envVars, ok := config.Envs[env]; ok {
		for key, value := range envVars {
			os.Setenv(key, value)
		}
	}

	return config.DefaultEnv, nil
}

func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Error loading .env file: %v\n", err)
	}
	defaultEnv, err := loadConfig()
	if err != nil {
		fmt.Printf("Error loading .localpost: %v\n", err)
	}
	if env == "" { // Only set if not already set (e.g., by flag)
		env = defaultEnv
	}
	if env == "" { // Fallback to dev if still unset
		env = "dev"
	}
	fmt.Fprintf(os.Stderr, "Using environment: %s\n", env)

	var rootCmd = &cobra.Command{
		Use:   "localpost",
		Short: "A CLI tool to manage and execute HTTP requests",
		Long:  `A tool to save and execute HTTP requests stored in a Git repository.`,
	}
	// Use StringP instead of StringVar to avoid overwriting env unless flag is provided
	rootCmd.PersistentFlags().StringP("env", "e", "", "Environment to use (e.g., dev, prod); defaults to .localpost or 'dev'")
	if flagEnv := rootCmd.PersistentFlags().Lookup("env").Value.String(); flagEnv != "" {
		env = flagEnv // Only override if flag is explicitly set
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
			if _, err := loadConfig(); err != nil { // Reload before each request
				fmt.Printf("Error loading .localpost: %v\n", err)
				os.Exit(1)
			}
			fmt.Fprintf(os.Stderr, "Request environment: %s\n", env)
			input := args[0]
			parts := strings.SplitN(input, " ", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: invalid input format, expected 'METHOD name' (e.g., GET users)\n")
				os.Exit(1)
			}
			method, name := parts[0], parts[1]
			if method != strings.ToUpper(method) {
				fmt.Printf("Error: method must be uppercase (e.g., GET, POST)\n")
				os.Exit(1)
			}

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
			req.EnvName = env // Pass the environment name to Request

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

	var setEnvCmd = &cobra.Command{
		Use:   "set-env [env]",
		Short: "Set the default environment",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			newEnv := args[0]
			configFilePath := filepath.Join(os.Getenv("HOME"), ".localpost")
			config := Config{
				DefaultEnv: "dev",
				Envs:       make(map[string]map[string]string),
			}
			if data, err := os.ReadFile(configFilePath); err == nil {
				if err := yaml.Unmarshal(data, &config); err != nil {
					fmt.Printf("Error parsing .localpost: %v\n", err)
					os.Exit(1)
				}
			}
			config.DefaultEnv = newEnv
			data, err := yaml.Marshal(&config)
			if err != nil {
				fmt.Printf("Error marshaling config: %v\n", err)
				os.Exit(1)
			}
			if err := os.WriteFile(configFilePath, data, 0644); err != nil {
				fmt.Printf("Error setting default env: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Default environment set to: %s\n", newEnv)
		},
	}

	rootCmd.AddCommand(greetCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(requestCmd)
	rootCmd.AddCommand(newRequestCmd)
	rootCmd.AddCommand(setEnvCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
