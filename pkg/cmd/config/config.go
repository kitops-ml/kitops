package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"gopkg.in/yaml.v3"
)

func setConfig(key, value, path string) error {

	configYamlPath := constants.ConfigYamlPath(path)

	configStruct, loadConfigErr := LoadConfigFileHelper(configYamlPath)
	if loadConfigErr != nil && !errors.Is(loadConfigErr, os.ErrNotExist) {
		return loadConfigErr
	}

	switch key {
	case "logLevel":
		configStruct.LogLevel = value
	case "progressBars":
		configStruct.ProgressBars = value
	case "verbosity":
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return err
		}
		configStruct.Verbosity = intValue
	default:
		return fmt.Errorf("invalid config key: %s", key)
	}

	if err := saveConfigFile(configStruct, configYamlPath); err != nil {
		return fmt.Errorf("failed to save setting: %w", err)
	}

	return nil
}

func getConfig(key, path string) (string, error) {
	configYamlPath := constants.ConfigYamlPath(path)

	configStruct, loadErr := LoadConfigFileHelper(configYamlPath)
	if loadErr != nil {
		return "", loadErr
	}

	switch key {
	case "logLevel":
		return configStruct.LogLevel, nil
	case "progressBars":
		return configStruct.ProgressBars, nil
	case "verbosity":
		stringValue := strconv.Itoa(configStruct.Verbosity)
		return stringValue, nil
	default:
		return "", fmt.Errorf("invalid config key: %s", key)
	}

}

func listConfig(path string) (Config, error) {
	configYamlPath := constants.ConfigYamlPath(path)

	configStruct, loadErr := LoadConfigFileHelper(configYamlPath)
	if loadErr != nil {
		return Config{}, loadErr
	}

	return configStruct, nil
}

func resetConfig(path string) error {
	configYamlPath := constants.ConfigYamlPath(path)

	if err := os.Remove(configYamlPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func LoadConfigFileHelper(configYamlPath string) (Config, error) {
	data, readErr := os.ReadFile(configYamlPath)
	var cfg Config

	if readErr != nil {
		if !errors.Is(readErr, os.ErrNotExist) {
			return cfg, readErr
		}
		return cfg, fmt.Errorf("config file does not exist: %w", readErr)
	}

	unmarshErr := yaml.Unmarshal(data, &cfg)
	if unmarshErr != nil {
		return cfg, fmt.Errorf("failed to unmarshal data: %w", unmarshErr)
	}

	return cfg, nil
}

func saveConfigFile(configStruct Config, configYamlPath string) error {
	yamlconfigStruct, marshErr := yaml.Marshal(configStruct)
	if marshErr != nil {
		return fmt.Errorf("failed to marshal data: %w", marshErr)
	}

	configDir := filepath.Dir(configYamlPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if writeErr := os.WriteFile(configYamlPath, yamlconfigStruct, 0644); writeErr != nil {
		return fmt.Errorf("failed to set to config file: %w", writeErr)
	}
	return nil
}
