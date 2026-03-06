package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"gopkg.in/yaml.v3"
)

func setConfig(key, value string) error {
	path, pathErr := constants.DefaultConfigPath()
	if pathErr != nil {
		return fmt.Errorf("Failed to get default path: %w", pathErr)
	}
	configMap := make(map[string]string)
	configYamlPath := constants.ConfigYamlPath(path)
	loadConfigErr := loadConfigFile(&configMap, path)

	if loadConfigErr != nil && !errors.Is(loadConfigErr, os.ErrNotExist) {
		return loadConfigErr
	}

	configMap[key] = value
	saveErr := saveSetConfig(configMap, configYamlPath)
	if saveErr != nil {
		return fmt.Errorf("Failed to save setting: %w", saveErr)
	}
	return nil
}

func saveSetConfig(configMap map[string]string, configYamlPath string) error {
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

func getConfig(key string) (string, error) {
	path, pathErr := constants.DefaultConfigPath()
	if pathErr != nil {
		return "", fmt.Errorf("Failed to get default path: %w", pathErr)
	}
	configMap := make(map[string]string)
	err := loadConfigFile(&configMap, path)
	if err != nil {
		return "", err
	}

	if val, ok := configMap[key]; ok {
		return val, nil
	} else {
		return "", fmt.Errorf("key %s does not exist", key)
	}
}

func loadConfigFile(configMap *map[string]string, path string) error {
	configYamlPath := constants.ConfigYamlPath(path)
	data, readErr := os.ReadFile(configYamlPath)

	if readErr == nil {
		unmarshErr := yaml.Unmarshal(data, configMap)
		if unmarshErr != nil {
			return fmt.Errorf("Failed to unmarshal data: %w", unmarshErr)
		}
	} else if !errors.Is(readErr, os.ErrNotExist) { // if there is an error (readErr != nil) that is not a missing file error
		return fmt.Errorf("Failed to read config file: %w", readErr)
	} else {
		return fmt.Errorf("Config %w", os.ErrNotExist)
	}

	return nil
}
