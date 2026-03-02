package config

const (
	configShortDesc = `Manage global configuration settings for the kirops CLI`
	configLongDesc  = `The config command allows you to view and modify persistent settings for the KitOps CLI.
	These settings are saved to a local configuration file,
	allowing you to establish default behaviors—such as default log levels or storage paths—
	without needing to pass flags manually on every command execution.
	Use the available subcommands to set, get, list, or reset your preferences.`

	configSetShortDesc = `Set a configuration value for KitOps`
	configSetLongDesc  = `Sets a specific configuration key to a given value in the local config.yaml file.
	If the configuration file or directory does not exist,
	it will be created automatically in the default cross-platform location.`
)
