package attest

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

type attestOptions struct {
	configHome string
	cosignArgs []string
}

func (opts *attestOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	//opts.cosignArgs = append(args, "verify")
	//cosignArgs :=
	opts.cosignArgs = append([]string{"attest"}, args...)
	fmt.Println(opts.cosignArgs)
	return nil
}

func AttestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "attest",
		Short:   "",
		Long:    "",
		Example: "",
		RunE:    runCommand(&attestOptions{}),
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

func runCommand(opts *attestOptions) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if err := opts.complete(cmd.Context(), args); err != nil {
			return output.Fatalf("Invalid arguments: %s", err)
		}

		err := RunAttest(cmd.Context(), opts)
		if err != nil {
			return output.Fatalf("Failed to attest: %s", err)
		}
		output.Infof("Attestation successful")
		return nil
	}
}
