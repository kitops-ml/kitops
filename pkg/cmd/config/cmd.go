package config

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/kitops-ml/kitops/pkg/lib/util"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

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
	if err := setConfig(args[0], args[1]); err != nil {
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
	val, err := getConfig(args[0])
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
	configMap, err := listConfig()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return output.Fatalf("%s", err)
	}

	keys := make([]string, 0, len(configMap))

	for k := range configMap {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	for _, k := range keys {
		fmt.Fprintf(output.GetOut(), "%s: %s\n", k, configMap[k])
	}

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

	return resetConfig()
}
