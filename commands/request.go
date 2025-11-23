package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moshe5745/localpost/util"
	"github.com/spf13/cobra"
)

func requestCompletionFunc(_ *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) >= 1 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	var requestPaths []string
	err := filepath.Walk(util.RequestsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
			relPath, err := filepath.Rel(util.RequestsDir, path)
			if err != nil {
				return err
			}
			relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
			requestPath := "/" + strings.TrimSuffix(relPath, ".yaml")
			if strings.HasPrefix(requestPath, "/"+strings.TrimPrefix(toComplete, "/")) {
				requestPaths = append(requestPaths, requestPath)
			}
		}
		return nil
	})
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return requestPaths, cobra.ShellCompDirectiveNoSpace
}

func RequestCmd() *cobra.Command {
	var verbose bool
	var inferSchema bool

	cmd := &cobra.Command{
		Use:     "request <path>",
		Aliases: []string{"r"},
		Short:   "Execute a request from a YAML file in the requests/ directory",
		Long: `Execute a request defined in a YAML file located in the requests/ directory.
Supports two formats:
  - Hierarchical: /path/to/dir/METHOD (e.g., /user/POST or /api/v1/auth/login/POST)
  - Flat: METHOD_name (e.g., GET_login, POST_user)
Use --infer-schema to generate a JTD schema from the response.
Use --verbose to show detailed request and response information.`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			requestPath := args[0]
			requestPath = strings.TrimPrefix(requestPath, "/")
			if requestPath == "" {
				fmt.Println("Error: request path cannot be empty")
				os.Exit(1)
			}

			validMethods := map[string]bool{
				"GET": true, "POST": true, "PUT": true, "DELETE": true,
				"PATCH": true, "HEAD": true, "OPTIONS": true, "TRACE": true,
			}

			var filePath string
			var err error

			// First, try the new hierarchical format: /path/to/dir/METHOD
			parts := strings.Split(requestPath, "/")
			if len(parts) >= 1 {
				method := parts[len(parts)-1]
				if validMethods[method] {
					// New format: path ends with HTTP method
					filePath = filepath.Join(util.RequestsDir, requestPath+".yaml")
				}
			}

			// If new format didn't work, try the old flat format: METHOD_name
			if filePath == "" {
				// Check if any valid HTTP method is a prefix of the requestPath
				for method := range validMethods {
					if strings.HasPrefix(strings.ToUpper(requestPath), method+"_") {
						// Old format: METHOD_name
						filePath = filepath.Join(util.RequestsDir, requestPath+".yaml")
						break
					}
				}
			}

			// If still no match, try to find the file by checking if it exists as-is
			if filePath == "" {
				testPath := filepath.Join(util.RequestsDir, requestPath+".yaml")
				if _, err := os.Stat(testPath); err == nil {
					filePath = testPath
				}
			}

			// If we still haven't found a valid file path, show error
			if filePath == "" {
				fmt.Printf("Error: could not find request file for '%s'.\n", requestPath)
				fmt.Printf("Expected formats:\n")
				fmt.Printf("  - Hierarchical: /path/to/dir/METHOD (e.g., /user/GET, /api/auth/POST)\n")
				fmt.Printf("  - Flat: METHOD_name (e.g., GET_login, POST_user)\n")
				fmt.Printf("Available requests:\n")
				
				// List available requests to help the user
				var requestPaths []string
				filepath.Walk(util.RequestsDir, func(path string, info os.FileInfo, err error) error {
					if !info.IsDir() && strings.HasSuffix(info.Name(), ".yaml") {
						relPath, err := filepath.Rel(util.RequestsDir, path)
						if err != nil {
							return err
						}
						relPath = strings.ReplaceAll(relPath, string(os.PathSeparator), "/")
						requestPaths = append(requestPaths, "  "+strings.TrimSuffix(relPath, ".yaml"))
					}
					return nil
				})
				if len(requestPaths) > 0 {
					fmt.Printf("%s\n", strings.Join(requestPaths, "\n"))
				} else {
					fmt.Printf("  (no requests found)\n")
				}
				os.Exit(1)
			}

			_, err = util.HandleRequest(filePath, verbose, inferSchema)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		},
		ValidArgsFunction: requestCompletionFunc,
	}

	cmd.Flags().BoolVar(&inferSchema, "infer-schema", true, "Generate a JTD schema from the response")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed request and response information")

	return cmd
}
