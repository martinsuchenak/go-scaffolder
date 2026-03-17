package postgen

import (
	"fmt"
	"os/exec"
)

func RunGoModTidy(dir string) (string, error) {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("go mod tidy failed: %w\nOutput: %s", err, output)
	}

	return string(output), nil
}
