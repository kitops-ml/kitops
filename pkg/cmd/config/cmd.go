package config

import (
	"encoding/json"

	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
	"strings"
)

type Config struct {
	Verbosity    int    `yaml:"verbosity" json:"verbosity"`
	LogLevel     string `yaml:"logLevel" json:"logLevel"`
	ProgressBars string `yaml:"progressBars" json:"progressBars"`
}

func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: configShortDesc,
		Long:  configLongDesc,
	}
	cmd.AddCommand(configSetCommand())
	cmd.AddCommand(configGetCommand())
	cmd.AddCommand(configListCommand())
	cmd.AddCommand(configResetCommand())

	return cmd
}

func configSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set KEY VALUE",
		Short: setConfigShortDesc,
		Long:  setConfigLongDesc,
		RunE:  runSetCommand,
	}
	cmd.Args = cobra.ExactArgs(2)

	return cmd
}

func runSetCommand(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	path, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("failed to retrieve config path from context")
	}
	if err := setConfig(args[0], args[1], path); err != nil {
		return output.Fatalf("%s", err)
	}
	return nil
}

func configGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get KEY",
		Short: getConfigShortDesc,
		Long:  getConfigLongDesc,
		RunE:  runGetCommand,
	}
	cmd.Args = cobra.ExactArgs(1)

	return cmd
}

func runGetCommand(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	path, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("failed to retrieve config path from context")
	}
	val, err := getConfig(args[0], path)
	if err != nil {
		return output.Fatalf("%s", err)
	}
	fmt.Fprintln(output.GetOut(), val)
	return nil
}

func configListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: listConfigShortDesc,
		Long:  listConfigLongDesc,
		RunE:  runListCommand,
	}
	cmd.Args = cobra.NoArgs

	return cmd
}

func runListCommand(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()
	path, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("failed to retrieve config path from context")
	}
	configStruct, err := listConfig(path)
	if err != nil {
		return output.Fatalf("%s", err)
	}

	list, jsonErr := json.MarshalIndent(configStruct, "", " ")

	if jsonErr != nil {
		return jsonErr
	}

	output.Infoln(string(list))

	return nil
}

func configResetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: resetConfigShortDesc,
		Long:  resetConfigLongDesc,
		RunE:  runResetCommand,
	}
	cmd.Args = cobra.NoArgs

	return cmd
}

func runResetCommand(cmd *cobra.Command, args []string) error {
	warning := "Warning: this action is destructive and cannot be undone.Proceed? (y/N): "

	choice, choiceErr := util.PromptForInput(warning, false)
	if choiceErr != nil {
		return choiceErr
	}

	if !strings.EqualFold(choice, "y") && !strings.EqualFold(choice, "yes") {
		return nil
	}

	ctx := cmd.Context()
	path, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("failed to retrieve config path from context")
	}

	return resetConfig(path)
}
