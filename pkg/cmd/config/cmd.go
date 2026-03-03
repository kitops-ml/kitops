package config

import (
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config {set, get, reset, list}",
		Short: configShortDesc,
		Long:  configLongDesc,
	}
	cmd.AddCommand(ConfigSetCommand())

	return cmd
}

func ConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set key value",
		Short: setConfigShortDesc,
		Long:  setConfigLongDesc,
		RunE:  runSetCommand(),
	}
	cmd.Args = cobra.ExactArgs(2)

	return cmd
}

func runSetCommand() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := setConfig(args[0], args[1])
		if err != nil{
			return output.Fatalf("%s", err)
		}
		return nil
	}
}


