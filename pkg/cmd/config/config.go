package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"gopkg.in/yaml.v3"
)

func setConfig(key, value string) error {
	configYamlPath, err := pathHelper()

	if err != nil {
		return err
	}

	configStruct, loadConfigErr := loadConfigFileHelper(configYamlPath)
	if loadConfigErr != nil && !errors.Is(loadConfigErr, os.ErrNotExist) {
		return loadConfigErr
	}

	switch key {
	case "logLevel":
		configStruct.LogLevel = value
	case "progressBars":
		configStruct.ProgressBars = value
	case "configHome":
		configStruct.ConfigHome = value
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

func getConfig(key string) (string, error) {
	configYamlPath, err := pathHelper()

	if err != nil {
		return "", err
	}

	configStruct, loadErr := loadConfigFileHelper(configYamlPath)
	if loadErr != nil {
		return "", loadErr
	}

	switch key {
	case "logLevel":
		return configStruct.LogLevel, nil
	case "progressBars":
		return configStruct.ProgressBars, nil
	case "configHome":
		return configStruct.ConfigHome, nil
	case "verbosity":
		stringValue := strconv.Itoa(configStruct.Verbosity)
		return stringValue, nil
	default:
		return "", fmt.Errorf("invalid config key: %s", key)
	}

}

func listConfig() (Config, error) {
	configYamlPath, err := pathHelper()

	if err != nil {
		return Config{}, err
	}

	configStruct, loadErr := loadConfigFileHelper(configYamlPath)
	if loadErr != nil {
		return Config{}, loadErr
	}

	return configStruct, nil
}

func resetConfig() error {
	configYamlPath, err := pathHelper()

	if err != nil {
		return err
	}

	if err := os.Remove(configYamlPath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

func loadConfigFileHelper(configYamlPath string) (Config, error) {
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
	if writeErr := os.WriteFile(configYamlPath, yamlconfigStruct, 0644); writeErr != nil {
		return fmt.Errorf("failed to set to config file: %w", writeErr)
	}
	return nil
}

func pathHelper() (string, error) {
	path, err := constants.DefaultConfigPath()
	if err != nil {
		return "", fmt.Errorf("failed to get default path: %w", err)
	}
	configYamlPath := constants.ConfigYamlPath(path)

	return configYamlPath, nil
}
