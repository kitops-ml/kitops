package config

import(
	"errors"
	"os"
	"path/filepath"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"gopkg.in/yaml.v3"
)

func setConfig(key, value string) error{
	path, pathErr := constants.DefaultConfigPath()
	if pathErr != nil {
		return fmt.Errorf("Failed to get default path: %w", pathErr)
	}

	configYamlPath := filepath.Join(path, "config.yaml")
	configMap := make(map[string]string)

	data, readErr := os.ReadFile(configYamlPath)
	// if there is error that is not a missing file error
	if readErr == nil{
		unmarshErr := yaml.Unmarshal(data, &configMap)
		if unmarshErr != nil {
			return fmt.Errorf("Failed to unmarshal data: %w", unmarshErr)
		}
	} else if !errors.Is(readErr, os.ErrNotExist) {
		return fmt.Errorf("File exists but failed to read: %w", readErr)
	}
	configMap[key] = value
	saveErr := saveConfig(configMap, configYamlPath)
	if saveErr != nil {
		return fmt.Errorf("Failed to save setting: %w", saveErr)
	}
	return nil
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