package verify

import (
	"context"
	"fmt"

	"github.com/kitops-ml/kitops/pkg/lib/completion"
	"github.com/kitops-ml/kitops/pkg/lib/constants"
	"github.com/kitops-ml/kitops/pkg/output"
	"github.com/spf13/cobra"
)

type verifyOptions struct {
	configHome string
	cosignArgs []string
}

func (opts *verifyOptions) complete(ctx context.Context, args []string) error {
	configHome, ok := ctx.Value(constants.ConfigKey{}).(string)
	if !ok {
		return fmt.Errorf("default config path not set on command context")
	}
	opts.configHome = configHome
	opts.cosignArgs = args

	return nil
}

func VerifyCommand() *cobra.Command {
	var verifySign, verifyAttestation bool

	cmd := &cobra.Command{
		Use:     "verify",
		Short:   "",
		Long:    "",
		Example: "",
		RunE:    runCommand([]verifyOptions{}, verifySign, verifyAttestation),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) >= 1 {
				return nil, cobra.ShellCompDirectiveNoFileComp
			}
			return completion.GetLocalModelKitsCompletion(cmd.Context(), toComplete), cobra.ShellCompDirectiveNoFileComp | cobra.ShellCompDirectiveNoSpace
		},
	}

	cmd.Flags().BoolVar(&verifySign, "verifysign", false, "If only modelkit signature needs to be verified")
	cmd.Flags().BoolVar(&verifyAttestation, "verifyattestation", false, "If only attestation needs to be verified")
	return cmd
}

func runCommand(opts []verifyOptions, verifySign, verifyAttestation bool) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		commands := []string{"verify", "verify-attestation"}

		if verifySign {
			opts = append(opts, verifyOptions{})
			args = append([]string{commands[0]}, args...)
			if err := opts[0].complete(cmd.Context(), args); err != nil {
				return output.Fatalf("Invalid arguments: %s", err)
			}
		} else if verifyAttestation {
			opts = append(opts, verifyOptions{})
			args = append([]string{commands[1]}, args...)
			if err := opts[0].complete(cmd.Context(), args); err != nil {
				return output.Fatalf("Invalid arguments: %s", err)
			}
		} else {
			for i := range 2 {
				opts = append(opts, verifyOptions{})
				args = append([]string{commands[i]}, args...)
				if err := opts[i].complete(cmd.Context(), args); err != nil {
					return output.Fatalf("Invalid arguments: %s", err)
				}
			}

		}

		for i := range len(opts) {
			err := RunVerify(cmd.Context(), opts[i])
			if err != nil {
				return output.Fatalf("Failed to %s: %s", commands[i], err)
			}
		}
		output.Infof("Modelkit signed")
		return nil
	}
}
