package config

import (
	"errors"
	"os"
	"path/filepath"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config {set, get, reset, list}",
		Short: configShortDesc,
		Long:  configLongDesc,
	}
	cmd.AddCommand(ConfigSetCommand())

	return cmd
}

func ConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set key value",
		Short: configSetShortDesc,
		Long:  configSetLongDesc,
		RunE:  runSetCommand(),
	}
	cmd.Args = cobra.ExactArgs(2)

	return cmd
}

func runSetCommand() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		path, pathErr := constants.DefaultConfigPath()
		if pathErr != nil {
			return output.Fatalf("Failed to get default path: %s", pathErr)
		}

		configYamlPath := filepath.Join(path, "config.yaml")
		configMap := make(map[string]string)

		data, readErr := os.ReadFile(configYamlPath)
		if errors.Is(readErr, os.ErrNotExist) {
			configMap[args[0]] = args[1]
			saveErr := saveConfig(configMap, configYamlPath)
			if saveErr != nil {
				return output.Fatalf("%s", saveErr)
			}
			return nil
		} else if readErr != nil { // if config file exists but permission to write is denied
			return output.Fatalf("Failed to read file: %s", readErr)
		}
		unmarshErr := yaml.Unmarshal(data, &configMap)
		if unmarshErr != nil {
			return output.Fatalf("Failed to unmarshal data: %s", unmarshErr)
		}
		configMap[args[0]] = args[1]
		saveErr := saveConfig(configMap, configYamlPath)
		if saveErr != nil {
			return output.Fatalf("%s", saveErr)
		}
		return nil
	}
}

func saveConfig(configMap map[string]string, configYamlPath string) error {
	yamlConfigMap, marshErr := yaml.Marshal(configMap)
	if marshErr != nil {
		return fmt.Errorf("Failed to marshal data: %w", marshErr)
	}
	writeErr := os.WriteFile(configYamlPath, yamlConfigMap, 0644)
	if writeErr != nil {
		return fmt.Errorf("Failed to set to config file: %w", writeErr)
	}
	return nil
}
