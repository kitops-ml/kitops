package verify

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func RunVerify(context context.Context, options verifyOptions) error {
	cmd := exec.CommandContext(context, "cosign", options.cosignArgs...)

	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("%s failed %s", options.cosignArgs[0], err)
	}
	return nil
}
