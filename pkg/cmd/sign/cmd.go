package sign

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

type signOptions struct {
	configHome string
	cosignArgs []string
}

func SignCommand() *cobra.Command {

	cmd := &cobra.Command{
		Use:     "sign",
		Short:   "",
		Long:    "",
		Example: "",
		RunE:    runCommand(&signOptions{}),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
		DisableFlagParsing: true,
	}

	return cmd
}

func (opts *signOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.cosignArgs = append([]string{"sign"}, args...)
	fmt.Println(opts.cosignArgs)
	return nil
}

func runCommand(opts *signOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		err := RunSign(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to sign: %s", err)
		}
		output.Infof("Modelkit signed")
		return nil
	}
}
