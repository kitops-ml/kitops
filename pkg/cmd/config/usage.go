package config

const (
	configShortDesc = `Manage global configuration settings for the kitops CLI`
	configLongDesc  = `The config command allows you to view and modify persistent settings for the Kitops CLI.
	These settings are saved to a local configuration file,
	allowing you to establish default behaviors—such as default log levels or storage paths—
	without needing to pass flags manually on every command execution.
	Use the available subcommands to set, get, list, or reset your preferences.`

	setConfigShortDesc = `Set a configuration value for Kitops`
	setConfigLongDesc  = `Sets a specific configuration key to a given value in the local config.yaml file.
	If the configuration file or directory does not exist,
	it will be created automatically in the default cross-platform location.`

	getConfigShortDesc = `Retrieve the value of a configuration key`
	getConfigLongDesc  = `Retrieves the currently set value for a specific configuration key from the KitOps configuration file. 
	If the key exists, its value will be printed to standard output. If the key is not found, the command will return an error.`

	listConfigShortDesc = `List all saved configuration settings`
	listConfigLongDesc  = `Prints a complete, alphabetically sorted list of all key-value pairs currently stored in the config.yaml file. 
	If no configuration file has been created yet, the command will exit quietly without outputting any text.`
)
