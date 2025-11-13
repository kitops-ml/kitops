package sign

import (
	"context"
	"fmt"
	"os/exec"
)

func RunSign(ctx context.Context, options *signOptions) error {
	cmd := exec.CommandContext(ctx, "cosign", options.cosignArgs...)

	cmd.Stdin = nil
	cmd.Stderr = nil
	cmd.Stderr = nil

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("signing failed %s", err)
	}
	return nil
}
