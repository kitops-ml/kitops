package config

import (
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

func ConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: configShortDesc,
		Long:  configLongDesc,
	}
	cmd.AddCommand(ConfigSetCommand())
	cmd.AddCommand(ConfigGetCommand())

	return cmd
}

func ConfigSetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set KEY VALUE",
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
		if err != nil {
			return output.Fatalf("%s", err)
		}
		return nil
	}
}

func ConfigGetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get KEY",
		Short: getConfigShortDesc,
		Long:  getConfigLongDesc,
		RunE:  runGetCommand(),
	}
	cmd.Args = cobra.ExactArgs(1)

	return cmd
}

func runGetCommand() func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		val, err := getConfig(args[0])
		if err != nil {
			return output.Fatalf("%s", err)
		}
		output.Infoln(val)
		return nil
	}
}
