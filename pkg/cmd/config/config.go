package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"gopkg.in/yaml.v3"
)

func setConfig(key, value string) error {
	path, configYamlPath, err := pathHelper()

	if err != nil {
		return err
	}

	configMap, loadConfigErr := loadConfigFileHelper(path)
	if loadConfigErr != nil {
		if errors.Is(loadConfigErr, os.ErrNotExist) {
			configMap = make(map[string]string)
		} else {
			return loadConfigErr
		}
	}

	configMap[key] = value
	saveErr := saveConfigFile(configMap, configYamlPath)
	if saveErr != nil {
		return fmt.Errorf("failed to save setting: %w", saveErr)
	}
	return nil
}

func saveConfigFile(configMap map[string]string, configYamlPath string) error {
	yamlConfigMap, marshErr := yaml.Marshal(configMap)
	if marshErr != nil {
		return fmt.Errorf("failed to marshal data: %w", marshErr)
	}
	writeErr := os.WriteFile(configYamlPath, yamlConfigMap, 0644)
	if writeErr != nil {
		return fmt.Errorf("failed to set to config file: %w", writeErr)
	}
	return nil
}

func getConfig(key string) (string, error) {
	path, pathErr := constants.DefaultConfigPath()
	if pathErr != nil {
		return "", fmt.Errorf("failed to get default path: %w", pathErr)
	}
	configMap, err := loadConfigFileHelper(path)
	if err != nil {
		return "", err
	}

	if val, ok := configMap[key]; ok {
		return val, nil
	} else {
		return "", fmt.Errorf("key %s does not exist", key)
	}
}

func listConfig() (map[string]string, error) {
	path, pathErr := constants.DefaultConfigPath()
	if pathErr != nil {
		return nil, fmt.Errorf("failed to get default path: %w", pathErr)
	}
	configMap, loadErr := loadConfigFileHelper(path)
	if loadErr != nil {
		return nil, loadErr
	}
	return configMap, nil
}

func resetConfig() error {
	path, configYamlPath, err := pathHelper()

	if err != nil {
		return err
	}

	configMap, loadErr := loadConfigFileHelper(path)
	if loadErr != nil {
		if errors.Is(loadErr, os.ErrNotExist) {
			// because resetting a non-existent config file is not really an error
			return nil
		}
		return loadErr
	}
	//clear map
	configMap = make(map[string]string)

	saveErr := saveConfigFile(configMap, configYamlPath)
	if saveErr != nil {
		return fmt.Errorf("failed to save setting: %w", saveErr)
	}

	return nil
}

func loadConfigFileHelper(path string) (map[string]string, error) {
	configYamlPath := constants.ConfigYamlPath(path)
	data, readErr := os.ReadFile(configYamlPath)

	if readErr != nil {
		if !errors.Is(readErr, os.ErrNotExist) {
			return nil, fmt.Errorf("failed to read config file: %w", readErr)
		}
		return nil, fmt.Errorf("config file does not exist: %w", os.ErrNotExist)
	}

	configMap := make(map[string]string)
	unmarshErr := yaml.Unmarshal(data, &configMap)
	if unmarshErr != nil {
		return nil, fmt.Errorf("failed to unmarshal data: %w", unmarshErr)
	}

	return configMap, nil
}

func pathHelper() (string, string, error) {
	path, pathErr := constants.DefaultConfigPath()
	if pathErr != nil {
		return "", "", fmt.Errorf("failed to get default path: %w", pathErr)
	}
	configYamlPath := constants.ConfigYamlPath(path)

	return path, configYamlPath, nil
}
