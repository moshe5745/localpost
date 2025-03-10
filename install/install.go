package completion

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func DetectShell() (string, string) {
	shellPath := os.Getenv("SHELL")
	if shellPath != "" {
		switch filepath.Base(shellPath) {
		case "bash":
			return "bash", filepath.Join(os.Getenv("HOME"), ".bashrc")
		case "zsh":
			return "zsh", filepath.Join(os.Getenv("HOME"), ".zshrc")
		case "fish":
			return "fish", filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions", "localpost.fish")
		}
	}
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", os.Getppid()), "-o", "comm=")
	output, err := cmd.Output()
	if err == nil {
		shell := strings.TrimSpace(string(output))
		switch shell {
		case "bash":
			return "bash", filepath.Join(os.Getenv("HOME"), ".bashrc")
		case "zsh":
			return "zsh", filepath.Join(os.Getenv("HOME"), ".zshrc")
		case "fish":
			return "fish", filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions", "localpost.fish")
		}
	}
	return "bash", filepath.Join(os.Getenv("HOME"), ".bashrc")
}

func InstallCompletion(rootCmd *cobra.Command) error {
	shell, configFile := DetectShell()
	fmt.Printf("Detected shell: %s\n", shell)

	var completionFile string
	switch shell {
	case "bash":
		completionFile = filepath.Join(os.Getenv("HOME"), ".localpost-install.bash")
	case "zsh":
		completionFile = filepath.Join(os.Getenv("HOME"), ".localpost-install.zsh")
	case "fish":
		completionFile = filepath.Join(os.Getenv("HOME"), ".config", "fish", "completions", "localpost.fish")
	default:
		completionFile = filepath.Join(os.Getenv("HOME"), ".localpost-install.bash")
	}

	f, err := os.Create(completionFile)
	if err != nil {
		return fmt.Errorf("error creating install file: %v", err)
	}
	defer f.Close()

	switch shell {
	case "bash":
		err = rootCmd.GenBashCompletion(f)
	case "zsh":
		_, err = f.WriteString(`# Initialize Zsh install system
autoload -U compinit && compinit

# Enable menu selection for localpost
zstyle ':install:*:*:localpost:*' menu yes select
zstyle ':install:*:*:localpost:*' list-colors 'di=34:ln=35:so=32:pi=33:ex=31:bd=46;34:cd=43;34:su=41;30:sg=46;30:tw=42;30:ow=43;30'

`)
		if err != nil {
			return fmt.Errorf("error writing zsh header: %v", err)
		}
		err = rootCmd.GenZshCompletion(f)
	case "fish":
		err = rootCmd.GenFishCompletion(f, true)
	default:
		err = rootCmd.GenBashCompletion(f)
	}
	if err != nil {
		return fmt.Errorf("error generating install: %v", err)
	}

	if shell == "fish" {
		fmt.Printf("Completion installed at %s.\nRun 'fish -c \"fish_update_completions\"' or restart your shell to enable.\n", completionFile)
		return nil
	}

	sourceLine := fmt.Sprintf("source %s", completionFile)
	configContent, err := os.ReadFile(configFile)
	configStr := string(configContent)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("error reading config file: %v", err)
	}

	needsUpdate := false
	if !strings.Contains(configStr, sourceLine) {
		needsUpdate = true
	}

	if shell == "zsh" && !strings.Contains(configStr, "compinit") {
		needsUpdate = true
	}

	if needsUpdate {
		fConfig, err := os.OpenFile(configFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error opening config file: %v", err)
		}
		defer fConfig.Close()

		if shell == "zsh" && !strings.Contains(string(configContent), "compinit") {
			if _, err := fConfig.WriteString("\n# Localpost CLI install setup\n" +
				"autoload -U compinit && compinit\n" +
				sourceLine + "\n"); err != nil {
				return fmt.Errorf("error writing to config file: %v", err)
			}
		} else if !strings.Contains(string(configContent), sourceLine) {
			if _, err := fConfig.WriteString("\n# Localpost CLI install\n" + sourceLine + "\n"); err != nil {
				return fmt.Errorf("error writing to config file: %v", err)
			}
		}
		fmt.Printf("Updated %s with install sourcing\n", configFile)
	} else {
		fmt.Printf("Completion sourcing already present in %s\n", configFile)
	}

	fmt.Printf("Completion installed at %s.\nTo enable now, run: source %s\nOr restart your shell.\n", completionFile, configFile)
	return nil
}
