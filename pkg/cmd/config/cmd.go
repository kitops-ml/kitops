package config

import (
	"errors"
	"fmt"
	"os"
	"sort"

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
	err := setConfig(args[0], args[1])
	if err != nil {
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
