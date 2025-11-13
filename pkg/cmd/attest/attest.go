package attest

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func RunAttest(context context.Context, options *attestOptions) any {
	cmd := exec.CommandContext(context, "cosign", options.cosignArgs...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("attestation failed %s", err)
	}
	return nil
}
