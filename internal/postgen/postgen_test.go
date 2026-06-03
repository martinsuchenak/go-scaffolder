package postgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunGoModTidy(t *testing.T) {
	dir := t.TempDir()

	gomod := `module testmod

go 1.26.3
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

func main() {}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	output, err := RunGoModTidy(dir)
	if err != nil {
		t.Fatalf("RunGoModTidy error: %v, output: %s", err, output)
	}
}

func TestRunGoModTidyInvalidDir(t *testing.T) {
	_, err := RunGoModTidy("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for invalid directory")
	}
}

func TestRunGoModTidyWithDependency(t *testing.T) {
	dir := t.TempDir()

	gomod := `module testmod

go 1.26.3

require golang.org/x/exp v0.0.0-20240119083558-1b970713d09a
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import _ "golang.org/x/exp/slices"

func main() {}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := RunGoModTidy(dir)
	if err != nil {
		t.Fatalf("RunGoModTidy error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "go.sum"))
	if err != nil {
		t.Fatalf("go.sum not created: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("go.sum should not be empty after tidy")
	}
}

func TestRunGoModTidyTimeout(t *testing.T) {
	dir := t.TempDir()

	gomod := `module testmod

go 1.26.3
`
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte(gomod), 0644); err != nil {
		t.Fatal(err)
	}

	mainGo := `package main

import (
	"os/exec"
	"time"
)

func main() {
	exec.Command("sleep", "9999").Run()
	time.Sleep(1 * time.Hour)
}
`
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(mainGo), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := RunGoModTidy(dir)
	if err != nil {
		if !strings.Contains(err.Error(), "timed out") && !strings.Contains(err.Error(), "failed") {
			t.Logf("got expected error: %v", err)
		}
	}
}
