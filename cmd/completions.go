package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

func installShellCompletions(rootCmd *cobra.Command, opts *rootOptions) error {
	if runtime.GOOS != "linux" && runtime.GOOS != "darwin" {
		return nil
	}

	configHome, err := getConfigHome(opts)
	if err != nil {
		return err
	}

	completionsDir := filepath.Join(configHome, "completions")
	err = os.MkdirAll(completionsDir, 0755)
	if err != nil {
		return err
	}

	shell := detectShell()
	if shell == "" {
		return fmt.Errorf("failed to detect shell, is $SHELL set?")
	}

	completionsFilePath := filepath.Join(completionsDir, "kitops."+shell)

	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	shellRcPath := ""
	switch shell {
	case "bash":
		shellRcPath = filepath.Join(userHomeDir, ".bashrc")
		if err := rootCmd.GenBashCompletionFileV2(completionsFilePath, true); err != nil {
			return err
		}
	case "zsh":
		shellRcPath = filepath.Join(userHomeDir, ".zshrc")
		if err := rootCmd.GenZshCompletionFile(completionsFilePath); err != nil {
			return err
		}
	case "fish":
		shellRcPath = filepath.Join(userHomeDir, ".config", "fish", "config.fish")
		if err := rootCmd.GenFishCompletionFile(completionsFilePath, true); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}

	if err := writeActivationScriptIfMissing(shellRcPath, completionsFilePath); err != nil {
		return err
	}

	return nil
}

func writeActivationScriptIfMissing(pathOfShellRC, pathOfCompletionFile string) error {
	content, err := os.ReadFile(pathOfShellRC)
	if err != nil {
		return err
	}
	sourceCmd := "source " + pathOfCompletionFile
	if strings.Contains(string(content), sourceCmd) {
		return nil // already present
	}

	shellRC, err := os.OpenFile(pathOfShellRC, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer shellRC.Close()

	_, err = shellRC.WriteString("\n# Kitops completions\nsource " + pathOfCompletionFile + "\n")

	if err == nil {
		output.Infof("Shell completions for Kitops have been installed. To activate them in your current session, run:\n\n\tsource %s\n\nOr restart your shell.\n", pathOfCompletionFile)
	}
	return err
}

func detectShell() string {
	shell := os.Getenv("SHELL")
	parts := strings.Split(shell, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}
