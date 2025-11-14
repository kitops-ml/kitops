package sign

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func RunSign(ctx context.Context, options *signOptions) error {
	cmd := exec.CommandContext(ctx, "cosign", options.cosignArgs...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("signing failed %s", err)
	}
	return nil
}
