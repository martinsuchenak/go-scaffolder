package postgen

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

const defaultTimeout = 5 * time.Minute

func RunGoModTidy(dir string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "mod", "tidy")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return string(output), fmt.Errorf("go mod tidy timed out after %v: %w", defaultTimeout, ctx.Err())
		}
		return string(output), fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}
